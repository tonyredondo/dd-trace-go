// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024 Datadog, Inc.

package tracer

import (
	"fmt"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/ext"
	"gopkg.in/DataDog/dd-trace-go.v1/internal"
	"gopkg.in/DataDog/dd-trace-go.v1/internal/globalconfig"
	"gopkg.in/DataDog/dd-trace-go.v1/internal/log"
	"gopkg.in/DataDog/dd-trace-go.v1/internal/namingschema"
	"gopkg.in/DataDog/dd-trace-go.v1/internal/traceprof"
	"net/url"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"
	"time"
)

// newCiVisibilityConfig creates and configures a new tracer configuration instance for CI Visibility.
// It sets default values, reads environment variables, and applies user-provided options.
//
// Parameters:
//
//	opts - A variadic list of StartOption functions to customize the configuration.
//
// Returns:
//
//	A pointer to the configured config instance.
func newCiVisibilityConfig(opts ...StartOption) *config {
	c := new(config)
	c.sampler = NewAllSampler()
	c.ciVisibilityEnabled = true
	c.httpClientTimeout = time.Second * 45 // 45 seconds
	c.logStartup = false                   // if we are in CI Visibility mode we don't log the startup to stdout to avoid polluting the output

	var err error
	c.hostname, err = os.Hostname()
	if err != nil {
		log.Warn("unable to look up hostname: %v", err)
	}

	if v := os.Getenv("DD_ENV"); v != "" {
		c.env = v
	}
	if v := os.Getenv("DD_TRACE_FEATURES"); v != "" {
		WithFeatureFlags(strings.FieldsFunc(v, func(r rune) bool {
			return r == ',' || r == ' '
		})...)(c)
	}
	if v := os.Getenv("DD_SERVICE"); v != "" {
		c.serviceName = v
		globalconfig.SetServiceName(v)
	}
	if ver := os.Getenv("DD_VERSION"); ver != "" {
		c.version = ver
	}
	if v := os.Getenv("DD_SERVICE_MAPPING"); v != "" {
		internal.ForEachStringTag(v, func(key, val string) { WithServiceMapping(key, val)(c) })
	}
	c.headerAsTags = newDynamicConfig("trace_header_tags", nil, setHeaderTags, equalSlice[string])
	if v := os.Getenv("DD_TRACE_HEADER_TAGS"); v != "" {
		WithHeaderTags(strings.Split(v, ","))(c)
	}
	if v := os.Getenv("DD_TAGS"); v != "" {
		tags := internal.ParseTagString(v)
		internal.CleanGitMetadataTags(tags)
		for key, val := range tags {
			WithGlobalTag(key, val)(c)
		}
	}

	c.debug = internal.BoolEnv("DD_TRACE_DEBUG", false)
	c.enabled = newDynamicConfig("tracing_enabled", internal.BoolEnv("DD_TRACE_ENABLED", true), func(b bool) bool { return true }, equal[bool])
	c.profilerEndpoints = internal.BoolEnv(traceprof.EndpointEnvVar, true)
	c.profilerHotspots = internal.BoolEnv(traceprof.CodeHotspotsEnvVar, true)
	c.enableHostnameDetection = internal.BoolEnv("DD_CLIENT_HOSTNAME_ENABLED", true)
	c.debugAbandonedSpans = internal.BoolEnv("DD_TRACE_DEBUG_ABANDONED_SPANS", false)
	if c.debugAbandonedSpans {
		c.spanTimeout = internal.DurationEnv("DD_TRACE_ABANDONED_SPAN_TIMEOUT", 10*time.Minute)
	}
	c.statsComputationEnabled = internal.BoolEnv("DD_TRACE_STATS_COMPUTATION_ENABLED", false)
	c.dataStreamsMonitoringEnabled = internal.BoolEnv("DD_DATA_STREAMS_ENABLED", false)
	c.partialFlushEnabled = internal.BoolEnv("DD_TRACE_PARTIAL_FLUSH_ENABLED", false)
	c.partialFlushMinSpans = internal.IntEnv("DD_TRACE_PARTIAL_FLUSH_MIN_SPANS", partialFlushMinSpansDefault)
	if c.partialFlushMinSpans <= 0 {
		log.Warn("DD_TRACE_PARTIAL_FLUSH_MIN_SPANS=%d is not a valid value, setting to default %d", c.partialFlushMinSpans, partialFlushMinSpansDefault)
		c.partialFlushMinSpans = partialFlushMinSpansDefault
	} else if c.partialFlushMinSpans >= traceMaxSize {
		log.Warn("DD_TRACE_PARTIAL_FLUSH_MIN_SPANS=%d is above the max number of spans that can be kept in memory for a single trace (%d spans), so partial flushing will never trigger, setting to default %d", c.partialFlushMinSpans, traceMaxSize, partialFlushMinSpansDefault)
		c.partialFlushMinSpans = partialFlushMinSpansDefault
	}
	// TODO(partialFlush): consider logging a warning if DD_TRACE_PARTIAL_FLUSH_MIN_SPANS
	// is set, but DD_TRACE_PARTIAL_FLUSH_ENABLED is not true. Or just assume it should be enabled
	// if it's explicitly set, and don't require both variables to be configured.

	c.dynamicInstrumentationEnabled = internal.BoolEnv("DD_DYNAMIC_INSTRUMENTATION_ENABLED", false)

	schemaVersionStr := os.Getenv("DD_TRACE_SPAN_ATTRIBUTE_SCHEMA")
	if v, ok := namingschema.ParseVersion(schemaVersionStr); ok {
		namingschema.SetVersion(v)
		c.spanAttributeSchemaVersion = int(v)
	} else {
		v := namingschema.SetDefaultVersion()
		c.spanAttributeSchemaVersion = int(v)
		log.Warn("DD_TRACE_SPAN_ATTRIBUTE_SCHEMA=%s is not a valid value, setting to default of v%d", schemaVersionStr, v)
	}
	// Allow DD_TRACE_SPAN_ATTRIBUTE_SCHEMA=v0 users to disable default integration (contrib AKA v0) service names.
	// These default service names are always disabled for v1 onwards.
	namingschema.SetUseGlobalServiceName(internal.BoolEnv("DD_TRACE_REMOVE_INTEGRATION_SERVICE_NAMES_ENABLED", false))

	// peer.service tag default calculation is enabled by default if using attribute schema >= 1
	c.peerServiceDefaultsEnabled = true
	if c.spanAttributeSchemaVersion == int(namingschema.SchemaV0) {
		c.peerServiceDefaultsEnabled = internal.BoolEnv("DD_TRACE_PEER_SERVICE_DEFAULTS_ENABLED", false)
	}
	c.peerServiceMappings = make(map[string]string)
	if v := os.Getenv("DD_TRACE_PEER_SERVICE_MAPPING"); v != "" {
		internal.ForEachStringTag(v, func(key, val string) { c.peerServiceMappings[key] = val })
	}

	for _, fn := range opts {
		fn(c)
	}

	if c.agentURL == nil {
		c.agentURL = resolveAgentAddr()
		if agentUrl := internal.AgentURLFromEnv(); agentUrl != nil {
			c.agentURL = agentUrl
		}
	}
	if c.agentURL.Scheme == "unix" {
		// If we're connecting over UDS we can just rely on the agent to provide the hostname
		log.Debug("connecting to agent over unix, do not set hostname on any traces")
		c.enableHostnameDetection = false
		c.httpClient = udsClient(c.agentURL.Path, c.httpClientTimeout)
		c.agentURL = &url.URL{
			Scheme: "http",
			Host:   fmt.Sprintf("UDS_%s", strings.NewReplacer(":", "_", "/", "_", `\`, "_").Replace(c.agentURL.Path)),
		}
	} else if c.httpClient == nil {
		c.httpClient = defaultHTTPClient(c.httpClientTimeout)
	}
	WithGlobalTag(ext.RuntimeID, globalconfig.RuntimeID())(c)
	globalTags := c.globalTags.get()
	if c.env == "" {
		if v, ok := globalTags["env"]; ok {
			if e, ok := v.(string); ok {
				c.env = e
			}
		}
	}
	if c.version == "" {
		if v, ok := globalTags["version"]; ok {
			if ver, ok := v.(string); ok {
				c.version = ver
			}
		}
	}
	if c.serviceName == "" {
		if v, ok := globalTags["service"]; ok {
			if s, ok := v.(string); ok {
				c.serviceName = s
				globalconfig.SetServiceName(s)
			}
		} else {
			// There is not an explicit service set, default to binary name.
			// In this case, don't set a global service name so the contribs continue using their defaults.
			c.serviceName = filepath.Base(os.Args[0])
		}
	}

	c.transport = newCiVisibilityTransport(c)

	if c.propagator == nil {
		envKey := "DD_TRACE_X_DATADOG_TAGS_MAX_LENGTH"
		max := internal.IntEnv(envKey, defaultMaxTagsHeaderLen)
		if max < 0 {
			log.Warn("Invalid value %d for %s. Setting to 0.", max, envKey)
			max = 0
		}
		if max > maxPropagatedTagsLength {
			log.Warn("Invalid value %d for %s. Maximum allowed is %d. Setting to %d.", max, envKey, maxPropagatedTagsLength, maxPropagatedTagsLength)
			max = maxPropagatedTagsLength
		}
		c.propagator = NewPropagator(&PropagatorConfig{
			MaxTagsHeaderLen: max,
		})
	}
	if c.logger != nil {
		log.UseLogger(c.logger)
	}
	if c.debug {
		log.SetLevel(log.LevelDebug)
	}

	c.agent = loadAgentFeatures(true, c.agentURL, c.httpClient)
	info, ok := debug.ReadBuildInfo()
	if !ok {
		c.loadContribIntegrations([]*debug.Module{})
	} else {
		c.loadContribIntegrations(info.Deps)
	}
	// Re-initialize the globalTags config with the value constructed from the environment and start options
	// This allows persisting the initial value of globalTags for future resets and updates.
	c.initGlobalTags(c.globalTags.get())

	return c
}

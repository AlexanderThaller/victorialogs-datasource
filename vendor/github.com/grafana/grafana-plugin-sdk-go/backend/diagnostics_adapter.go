package backend

import (
	"bytes"
	"context"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/expfmt"

	"github.com/grafana/grafana-plugin-sdk-go/genproto/pluginv2"
)

// diagnosticsSDKAdapter adapter between low level plugin protocol and SDK interfaces.
type diagnosticsSDKAdapter struct {
	metricGatherer     prometheus.Gatherer
	checkHealthHandler CheckHealthHandler
}

func newDiagnosticsSDKAdapter(metricGatherer prometheus.Gatherer, checkHealthHandler CheckHealthHandler) *diagnosticsSDKAdapter {
	return &diagnosticsSDKAdapter{
		metricGatherer:     metricGatherer,
		checkHealthHandler: checkHealthHandler,
	}
}

func (a *diagnosticsSDKAdapter) CollectMetrics(_ context.Context, _ *pluginv2.CollectMetricsRequest) (*pluginv2.CollectMetricsResponse, error) {
	mfs, err := a.metricGatherer.Gather()
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	for _, mf := range mfs {
		_, err := expfmt.MetricFamilyToText(&buf, mf)
		if err != nil {
			return nil, err
		}
	}

	return &pluginv2.CollectMetricsResponse{
		Metrics: &pluginv2.CollectMetricsResponse_Payload{
			Prometheus: buf.Bytes(),
		},
	}, nil
}

func (a *diagnosticsSDKAdapter) CheckHealth(ctx context.Context, protoReq *pluginv2.CheckHealthRequest) (*pluginv2.CheckHealthResponse, error) {
	if a.checkHealthHandler != nil {
		parsedReq := FromProto().CheckHealthRequest(protoReq)
		resp, err := a.checkHealthHandler.CheckHealth(ctx, parsedReq)
		if err != nil {
			return nil, err
		}

		return ToProto().CheckHealthResponse(resp), nil
	}

	return &pluginv2.CheckHealthResponse{
		Status: pluginv2.CheckHealthResponse_OK,
	}, nil
}

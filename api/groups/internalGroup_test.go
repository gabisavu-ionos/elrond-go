package groups_test

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	apiErrors "github.com/ElrondNetwork/elrond-go/api/errors"
	"github.com/ElrondNetwork/elrond-go/api/groups"
	"github.com/ElrondNetwork/elrond-go/api/mock"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRawBlockGroup(t *testing.T) {
	t.Parallel()

	t.Run("nil facade", func(t *testing.T) {
		hg, err := groups.NewInternalBlockGroup(nil)
		require.True(t, errors.Is(err, apiErrors.ErrNilFacadeHandler))
		require.Nil(t, hg)
	})

	t.Run("should work", func(t *testing.T) {
		hg, err := groups.NewInternalBlockGroup(&mock.FacadeStub{})
		require.NoError(t, err)
		require.NotNil(t, hg)
	})
}

func TestGetRawMetaBlockByNonce_EmptyNonceUrlParameterShouldErr(t *testing.T) {
	t.Parallel()

	facade := mock.FacadeStub{
		GetInternalMetaBlockByNonceCalled: func(_ common.OutportFormat, _ uint64) (interface{}, error) {
			return []byte{}, nil
		},
	}

	blockGroup, err := groups.NewInternalBlockGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(blockGroup, "raw", getRawBlockRoutesConfig())

	req, _ := http.NewRequest("GET", "/internal/raw/metablock/by-nonce", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	response := blockResponse{}
	loadResponse(resp.Body, &response)
	assert.Equal(t, http.StatusNotFound, resp.Code)
}

func TestGetRawMetaBlockByNonce_InvalidNonceShouldErr(t *testing.T) {
	t.Parallel()

	facade := mock.FacadeStub{
		GetInternalMetaBlockByNonceCalled: func(_ common.OutportFormat, _ uint64) (interface{}, error) {
			return []byte{}, nil
		},
	}

	blockGroup, err := groups.NewInternalBlockGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(blockGroup, "internal", getRawBlockRoutesConfig())

	req, _ := http.NewRequest("GET", "/internal/raw/metablock/by-nonce/invalid", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	response := blockResponse{}
	loadResponse(resp.Body, &response)
	assert.Equal(t, http.StatusBadRequest, resp.Code)
}

func TestGetRawMetaBlockByNonce_ShouldWork(t *testing.T) {
	t.Parallel()

	expectedOutput := bytes.Repeat([]byte("1"), 10)

	facade := mock.FacadeStub{
		GetInternalMetaBlockByNonceCalled: func(_ common.OutportFormat, _ uint64) (interface{}, error) {
			return expectedOutput, nil
		},
	}

	blockGroup, err := groups.NewInternalBlockGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(blockGroup, "internal", getRawBlockRoutesConfig())

	req, _ := http.NewRequest("GET", "/internal/raw/metablock/by-nonce/15", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	response := rawBlockResponse{}
	loadResponse(resp.Body, &response)
	assert.Equal(t, http.StatusOK, resp.Code)

	assert.Equal(t, expectedOutput, response.Data.Block)
}

func TestGetRawMetaBlockByNonceMetaBlockCheck_ShouldWork(t *testing.T) {
	t.Parallel()

	expectedOutput := bytes.Repeat([]byte("1"), 10)

	facade := mock.FacadeStub{
		GetInternalMetaBlockByNonceCalled: func(_ common.OutportFormat, _ uint64) (interface{}, error) {
			return expectedOutput, nil
		},
	}

	blockGroup, err := groups.NewInternalBlockGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(blockGroup, "internal", getRawBlockRoutesConfig())

	req, _ := http.NewRequest("GET", "/internal/raw/metablock/by-nonce/15", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	response := rawBlockResponse{}
	loadResponse(resp.Body, &response)
	assert.Equal(t, http.StatusOK, resp.Code)

	assert.Equal(t, expectedOutput, response.Data.Block)
}

// ----------------- Shard Block ---------------

func TestGetRawShardBlockByNonce_EmptyNonceUrlParameterShouldErr(t *testing.T) {
	t.Parallel()

	facade := mock.FacadeStub{
		GetInternalMetaBlockByNonceCalled: func(_ common.OutportFormat, _ uint64) (interface{}, error) {
			return []byte{}, nil
		},
	}

	blockGroup, err := groups.NewInternalBlockGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(blockGroup, "raw", getRawBlockRoutesConfig())

	req, _ := http.NewRequest("GET", "/internal/raw/shardblock/by-nonce", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	response := blockResponse{}
	loadResponse(resp.Body, &response)
	assert.Equal(t, http.StatusNotFound, resp.Code)
}

func TestGetRawShardBlockByNonce_InvalidNonceShouldErr(t *testing.T) {
	t.Parallel()

	facade := mock.FacadeStub{
		GetInternalShardBlockByNonceCalled: func(_ common.OutportFormat, _ uint64) (interface{}, error) {
			return []byte{}, nil
		},
	}

	blockGroup, err := groups.NewInternalBlockGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(blockGroup, "internal", getRawBlockRoutesConfig())

	req, _ := http.NewRequest("GET", "/internal/raw/shardblock/by-nonce/invalid", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	response := rawBlockResponse{}
	loadResponse(resp.Body, &response)
	assert.Equal(t, http.StatusBadRequest, resp.Code)
}

func TestGetRawShardBlockByNonce_ShouldWork(t *testing.T) {
	t.Parallel()

	expectedOutput := bytes.Repeat([]byte("1"), 10)

	facade := mock.FacadeStub{
		GetInternalShardBlockByNonceCalled: func(_ common.OutportFormat, _ uint64) (interface{}, error) {
			return expectedOutput, nil
		},
	}

	blockGroup, err := groups.NewInternalBlockGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(blockGroup, "internal", getRawBlockRoutesConfig())

	req, _ := http.NewRequest("GET", "/internal/raw/shardblock/by-nonce/15", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	response := rawBlockResponse{}
	loadResponse(resp.Body, &response)
	assert.Equal(t, http.StatusOK, resp.Code)

	assert.Equal(t, expectedOutput, response.Data.Block)
}

func TestGetRawShardBlockByNonceMetaBlockCheck_ShouldWork(t *testing.T) {
	t.Parallel()

	expectedOutput := bytes.Repeat([]byte("1"), 10)

	facade := mock.FacadeStub{
		GetInternalShardBlockByNonceCalled: func(_ common.OutportFormat, _ uint64) (interface{}, error) {
			return expectedOutput, nil
		},
	}

	blockGroup, err := groups.NewInternalBlockGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(blockGroup, "internal", getRawBlockRoutesConfig())

	req, _ := http.NewRequest("GET", "/internal/raw/shardblock/by-nonce/15", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	response := rawBlockResponse{}
	loadResponse(resp.Body, &response)
	assert.Equal(t, http.StatusOK, resp.Code)

	assert.Equal(t, expectedOutput, response.Data.Block)
}

// ---- MiniBlock

type rawMiniBlockResponseData struct {
	Block []byte `json:"miniblock"`
}

type rawMiniBlockResponse struct {
	Data  rawMiniBlockResponseData `json:"data"`
	Error string                   `json:"error"`
	Code  string                   `json:"code"`
}

func TestGetRawMiniBlockByHash_EmptyHashUrlParameterShouldErr(t *testing.T) {
	t.Parallel()

	facade := mock.FacadeStub{
		GetInternalMiniBlockByHashCalled: func(_ common.OutportFormat, _ string) (interface{}, error) {
			return []byte{}, nil
		},
	}

	blockGroup, err := groups.NewInternalBlockGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(blockGroup, "raw", getRawBlockRoutesConfig())

	req, _ := http.NewRequest("GET", "/internal/raw/miniblock/by-hash", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	response := rawBlockResponse{}
	loadResponse(resp.Body, &response)
	assert.Equal(t, http.StatusNotFound, resp.Code)
}

func TestGetRawMiniBlockByHash_ShouldWork(t *testing.T) {
	t.Parallel()

	expectedOutput := bytes.Repeat([]byte("1"), 10)

	facade := mock.FacadeStub{
		GetInternalMiniBlockByHashCalled: func(_ common.OutportFormat, _ string) (interface{}, error) {
			return expectedOutput, nil
		},
	}

	blockGroup, err := groups.NewInternalBlockGroup(&facade)
	require.NoError(t, err)

	ws := startWebServer(blockGroup, "internal", getRawBlockRoutesConfig())

	req, _ := http.NewRequest("GET", "/internal/raw/miniblock/by-hash/dummyhash", nil)
	resp := httptest.NewRecorder()
	ws.ServeHTTP(resp, req)

	response := rawMiniBlockResponse{}
	loadResponse(resp.Body, &response)
	assert.Equal(t, http.StatusOK, resp.Code)

	assert.Equal(t, expectedOutput, response.Data.Block)
}

func getRawBlockRoutesConfig() config.ApiRoutesConfig {
	return config.ApiRoutesConfig{
		APIPackages: map[string]config.APIPackageConfig{
			"internal": {
				Routes: []config.RouteConfig{
					{Name: "/raw/metablock/by-nonce/:nonce", Open: true},
					{Name: "/raw/metablock/by-hash/:hash", Open: true},
					{Name: "/raw/metablock/by-round/:round", Open: true},
					{Name: "/raw/shardblock/by-nonce/:nonce", Open: true},
					{Name: "/raw/shardblock/by-hash/:hash", Open: true},
					{Name: "/raw/shardblock/by-round/:round", Open: true},
					{Name: "/raw/miniblock/by-hash/:hash", Open: true},
					{Name: "/json/metablock/by-nonce/:nonce", Open: true},
					{Name: "/json/metablock/by-hash/:hash", Open: true},
					{Name: "/json/metablock/by-round/:round", Open: true},
					{Name: "/json/shardblock/by-nonce/:nonce", Open: true},
					{Name: "/json/shardblock/by-hash/:hash", Open: true},
					{Name: "/json/shardblock/by-round/:round", Open: true},
					{Name: "/json/miniblock/by-hash/:hash", Open: true},
				},
			},
		},
	}
}

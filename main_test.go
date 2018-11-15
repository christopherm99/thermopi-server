package main

import (
	"encoding/json"
	"github.com/labstack/echo"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"testing"
)

func TestTargetHandlers(t *testing.T) {
	e := echo.New()
	query := `{"target":21,"persistent":false}`
	postRequest := httptest.NewRequest(http.MethodPost, "/target", strings.NewReader(query))
	postRecord := httptest.NewRecorder()
	postContext := e.NewContext(postRequest, postRecord)
	getRequest := httptest.NewRequest(http.MethodGet, "/target", nil)
	getRecord := httptest.NewRecorder()
	getContext := e.NewContext(getRequest, getRecord)
	if assert.NoError(t, postTarget(postContext)) {
		assert.Equal(t, http.StatusAccepted, postRecord.Code)
		if assert.NoError(t, getTarget(getContext)) {
			assert.Equal(t, http.StatusOK, getRecord.Code)
			assert.Equal(t, `{"value":21}`, getRecord.Body.String())
		}
	}
}

func TestSensorHandlers(t *testing.T) {
	e := echo.New()
	query1 := url.Values{}
	query2 := url.Values{}
	query1.Set("value", "10")
	query2.Set("value", "30")
	post1Request := httptest.NewRequest(http.MethodPost, "/sensors/Bedroom?"+query1.Encode(), nil)
	post2Request := httptest.NewRequest(http.MethodPost, "/sensors/Kitchen?"+query2.Encode(), nil)
	post1Record := httptest.NewRecorder()
	post2Record := httptest.NewRecorder()
	post1Context := e.NewContext(post1Request, post1Record)
	post2Context := e.NewContext(post2Request, post2Record)
	post1Context.SetPath("/sensors/:id")
	post1Context.SetParamNames("id")
	post1Context.SetParamValues("Bedroom")
	post2Context.SetPath("/sensors/:id")
	post2Context.SetParamNames("id")
	post2Context.SetParamValues("Kitchen")
	getRequest := httptest.NewRequest(http.MethodGet, "/sensors", nil)
	getRecord := httptest.NewRecorder()
	getContext := e.NewContext(getRequest, getRecord)
	if assert.NoError(t, postSensors(post1Context)) {
		assert.Equal(t, http.StatusAccepted, post1Record.Code)
		if assert.NoError(t, postSensors(post2Context)) {
			assert.Equal(t, http.StatusAccepted, post2Record.Code)
			if assert.NoError(t, getSensors(getContext)) {
				assert.Equal(t, http.StatusOK, getRecord.Code)
				var expect struct {
					Name  string `json:"name"`
					Value int    `json:"value"`
				}
				assert.NotNil(t, json.Unmarshal([]byte(`[{"name":"Bedroom","value":10},{"name":"Kitchen","value":30}]`), &expect))
				var actual struct {
					Name  string `json:"name"`
					Value int    `json:"value"`
				}
				assert.NotNil(t, json.Unmarshal(getRecord.Body.Bytes(), &actual))
				assert.True(t, reflect.DeepEqual(expect, actual))
			}
		}
	}
}

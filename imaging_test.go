package onvif

import (
	"context"
	"net/http/httptest"
	"testing"
)

// newImagingTestClient returns a Client whose imagingEndpoint points at server.
func newImagingTestClient(t *testing.T, server *httptest.Server) *Client {
	t.Helper()

	client, err := NewClient(server.URL, WithHTTPClient(server.Client()))
	if err != nil {
		t.Fatalf("NewClient() failed: %v", err)
	}
	client.imagingEndpoint = server.URL

	return client
}

func TestGetImagingSettings(t *testing.T) {
	body := `<GetImagingSettingsResponse>
        <ImagingSettings>
            <Brightness>50</Brightness>
            <Contrast>60</Contrast>
        </ImagingSettings>
    </GetImagingSettingsResponse>`
	client := newImagingTestClient(t, newSOAPTestServer(t, body))

	settings, err := client.GetImagingSettings(context.Background(), "videosource1")
	if err != nil {
		t.Fatalf("GetImagingSettings() error = %v", err)
	}
	if settings.Brightness == nil || *settings.Brightness != 50 {
		t.Errorf("GetImagingSettings() Brightness = %v, want 50", settings.Brightness)
	}
}

func TestSetImagingSettings(t *testing.T) {
	client := newImagingTestClient(t, newSOAPTestServer(t, soapAckOnlyResponse))

	brightness := 75.0
	err := client.SetImagingSettings(context.Background(), "videosource1", &ImagingSettings{Brightness: &brightness}, false)
	if err != nil {
		t.Errorf("SetImagingSettings() error = %v", err)
	}
}

func TestImagingMove(t *testing.T) {
	client := newImagingTestClient(t, newSOAPTestServer(t, soapAckOnlyResponse))

	// FocusMove is currently a placeholder struct in the client (imaging.go)
	// with no move-type fields yet, so there is nothing to populate here.
	err := client.Move(context.Background(), "videosource1", &FocusMove{})
	if err != nil {
		t.Errorf("Move() error = %v", err)
	}
}

func TestImagingGetOptions(t *testing.T) {
	body := `<GetOptionsResponse>
        <ImagingOptions>
            <Brightness><Min>0</Min><Max>100</Max></Brightness>
        </ImagingOptions>
    </GetOptionsResponse>`
	client := newImagingTestClient(t, newSOAPTestServer(t, body))

	options, err := client.GetOptions(context.Background(), "videosource1")
	if err != nil {
		t.Fatalf("GetOptions() error = %v", err)
	}
	if options.Brightness == nil || options.Brightness.Max != 100 {
		t.Errorf("GetOptions() Brightness = %+v, want Max=100", options.Brightness)
	}
}

func TestGetMoveOptions(t *testing.T) {
	body := `<GetMoveOptionsResponse>
        <MoveOptions>
            <Absolute>
                <Position><Min>0</Min><Max>1</Max></Position>
                <Speed><Min>0</Min><Max>1</Max></Speed>
            </Absolute>
        </MoveOptions>
    </GetMoveOptionsResponse>`
	client := newImagingTestClient(t, newSOAPTestServer(t, body))

	options, err := client.GetMoveOptions(context.Background(), "videosource1")
	if err != nil {
		t.Fatalf("GetMoveOptions() error = %v", err)
	}
	if options.Absolute == nil || options.Absolute.Position.Max != 1 {
		t.Errorf("GetMoveOptions() Absolute = %+v, want Position.Max=1", options.Absolute)
	}
}

func TestStopFocus(t *testing.T) {
	client := newImagingTestClient(t, newSOAPTestServer(t, soapAckOnlyResponse))

	if err := client.StopFocus(context.Background(), "videosource1"); err != nil {
		t.Errorf("StopFocus() error = %v", err)
	}
}

func TestGetImagingStatus(t *testing.T) {
	body := `<GetStatusResponse>
        <Status>
            <FocusStatus>
                <Position>0.42</Position>
                <MoveStatus>IDLE</MoveStatus>
            </FocusStatus>
        </Status>
    </GetStatusResponse>`
	client := newImagingTestClient(t, newSOAPTestServer(t, body))

	status, err := client.GetImagingStatus(context.Background(), "videosource1")
	if err != nil {
		t.Fatalf("GetImagingStatus() error = %v", err)
	}
	if status.FocusStatus == nil || status.FocusStatus.Position != 0.42 {
		t.Errorf("GetImagingStatus() FocusStatus = %+v, want Position=0.42", status.FocusStatus)
	}
}

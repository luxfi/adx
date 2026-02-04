package vast

import (
	"encoding/xml"
	"time"
)

// VAST 4.x Video Ad Serving Template
type VAST struct {
	XMLName xml.Name `xml:"VAST"`
	Version string   `xml:"version,attr"`
	Ads     []Ad     `xml:"Ad"`
}

// Ad represents a VAST advertisement
type Ad struct {
	ID       string   `xml:"id,attr"`
	Sequence int      `xml:"sequence,attr,omitempty"`
	InLine   *InLine  `xml:"InLine,omitempty"`
	Wrapper  *Wrapper `xml:"Wrapper,omitempty"`
}

// InLine contains all data to display the ad
type InLine struct {
	AdSystem    AdSystem     `xml:"AdSystem"`
	AdTitle     string       `xml:"AdTitle"`
	Description string       `xml:"Description,omitempty"`
	Advertiser  string       `xml:"Advertiser,omitempty"`
	Pricing     *Pricing     `xml:"Pricing,omitempty"`
	Survey      []string     `xml:"Survey,omitempty"`
	Error       []string     `xml:"Error,omitempty"`
	Impression  []Impression `xml:"Impression"`
	Creatives   Creatives    `xml:"Creatives"`
	Extensions  *Extensions  `xml:"Extensions,omitempty"`
}

// Wrapper points to another VAST response
type Wrapper struct {
	AdSystem       AdSystem     `xml:"AdSystem"`
	VASTAdTagURI   string       `xml:"VASTAdTagURI"`
	Error          []string     `xml:"Error,omitempty"`
	Impression     []Impression `xml:"Impression"`
	Creatives      Creatives    `xml:"Creatives,omitempty"`
	Extensions     *Extensions  `xml:"Extensions,omitempty"`
	FallbackOnNoAd bool         `xml:"fallbackOnNoAd,attr,omitempty"`
}

// AdSystem info
type AdSystem struct {
	Version string `xml:"version,attr,omitempty"`
	Name    string `xml:",chardata"`
}

// Pricing information
type Pricing struct {
	Model    string  `xml:"model,attr"`
	Currency string  `xml:"currency,attr"`
	Value    float64 `xml:",chardata"`
}

// Impression tracking pixel
type Impression struct {
	ID  string `xml:"id,attr,omitempty"`
	URL string `xml:",cdata"`
}

// Creatives container
type Creatives struct {
	Creative []Creative `xml:"Creative"`
}

// Creative element
type Creative struct {
	ID           string        `xml:"id,attr,omitempty"`
	AdID         string        `xml:"adId,attr,omitempty"`
	Sequence     int           `xml:"sequence,attr,omitempty"`
	Linear       *Linear       `xml:"Linear,omitempty"`
	NonLinearAds *NonLinearAds `xml:"NonLinearAds,omitempty"`
	CompanionAds *CompanionAds `xml:"CompanionAds,omitempty"`
}

// Linear video ad
type Linear struct {
	SkipOffset     string          `xml:"skipoffset,attr,omitempty"`
	Duration       string          `xml:"Duration"`
	AdParameters   *AdParameters   `xml:"AdParameters,omitempty"`
	MediaFiles     MediaFiles      `xml:"MediaFiles"`
	VideoClicks    *VideoClicks    `xml:"VideoClicks,omitempty"`
	TrackingEvents *TrackingEvents `xml:"TrackingEvents,omitempty"`
	Icons          *Icons          `xml:"Icons,omitempty"`
}

// MediaFiles container
type MediaFiles struct {
	MediaFile []MediaFile `xml:"MediaFile"`
	Mezzanine *Mezzanine  `xml:"Mezzanine,omitempty"`
}

// MediaFile represents a video file
type MediaFile struct {
	ID                  string `xml:"id,attr,omitempty"`
	Delivery            string `xml:"delivery,attr"`
	Type                string `xml:"type,attr"`
	Bitrate             int    `xml:"bitrate,attr,omitempty"`
	MinBitrate          int    `xml:"minBitrate,attr,omitempty"`
	MaxBitrate          int    `xml:"maxBitrate,attr,omitempty"`
	Width               int    `xml:"width,attr"`
	Height              int    `xml:"height,attr"`
	Scalable            bool   `xml:"scalable,attr,omitempty"`
	MaintainAspectRatio bool   `xml:"maintainAspectRatio,attr,omitempty"`
	Codec               string `xml:"codec,attr,omitempty"`
	APIFramework        string `xml:"apiFramework,attr,omitempty"`
	URL                 string `xml:",cdata"`
}

// Mezzanine file for high-quality source
type Mezzanine struct {
	ID       string `xml:"id,attr,omitempty"`
	Delivery string `xml:"delivery,attr"`
	Type     string `xml:"type,attr"`
	Width    int    `xml:"width,attr"`
	Height   int    `xml:"height,attr"`
	Codec    string `xml:"codec,attr,omitempty"`
	URL      string `xml:",cdata"`
}

// VideoClicks for clickthrough and tracking
type VideoClicks struct {
	ClickThrough  *ClickThrough   `xml:"ClickThrough,omitempty"`
	ClickTracking []ClickTracking `xml:"ClickTracking,omitempty"`
	CustomClick   []CustomClick   `xml:"CustomClick,omitempty"`
}

// ClickThrough URL
type ClickThrough struct {
	ID  string `xml:"id,attr,omitempty"`
	URL string `xml:",cdata"`
}

// ClickTracking URL
type ClickTracking struct {
	ID  string `xml:"id,attr,omitempty"`
	URL string `xml:",cdata"`
}

// CustomClick URL
type CustomClick struct {
	ID  string `xml:"id,attr,omitempty"`
	URL string `xml:",cdata"`
}

// AdPod for CTV ad breaks
type AdPod struct {
	ID            string
	MaxDuration   time.Duration
	MinDuration   time.Duration
	MaxAds        int
	MinAds        int
	AdBreakType   string // pre-roll, mid-roll, post-roll
	Ads           []Ad
	TotalDuration time.Duration
}

// CTVAdRequest for Connected TV specific requests
type CTVAdRequest struct {
	DeviceID      string
	AppID         string
	ContentID     string
	ContentGenre  string
	ContentRating string
	AdPodSlots    []AdPod
	UserAgent     string
	IP            string
	Lat           float64
	Lon           float64
	DNT           bool
	LMT           bool // Limit Ad Tracking
	COPPA         bool
}

// CTVAdResponse for Connected TV
type CTVAdResponse struct {
	RequestID string
	AdPods    []AdPod
	VAST      *VAST
	Errors    []string
}

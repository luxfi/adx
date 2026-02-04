package vast

import (
	"context"
	"time"
)

// Additional VAST types not in vast.go

// TrackingEvents container
type TrackingEvents struct {
	Tracking []Tracking `xml:"Tracking"`
}

// Tracking event
type Tracking struct {
	Event  string `xml:"event,attr"`
	Offset string `xml:"offset,attr,omitempty"`
	URL    string `xml:",cdata"`
}

// AdParameters for VPAID
type AdParameters struct {
	XMLEncoded bool   `xml:"xmlEncoded,attr,omitempty"`
	Parameters string `xml:",cdata"`
}

// Icons container
type Icons struct {
	Icon []Icon `xml:"Icon"`
}

// Icon element
type Icon struct {
	Program          string          `xml:"program,attr,omitempty"`
	Width            int             `xml:"width,attr"`
	Height           int             `xml:"height,attr"`
	XPosition        string          `xml:"xPosition,attr"`
	YPosition        string          `xml:"yPosition,attr"`
	Duration         string          `xml:"duration,attr,omitempty"`
	Offset           string          `xml:"offset,attr,omitempty"`
	APIFramework     string          `xml:"apiFramework,attr,omitempty"`
	StaticResource   *StaticResource `xml:"StaticResource,omitempty"`
	IFrameResource   string          `xml:"IFrameResource,omitempty"`
	HTMLResource     string          `xml:"HTMLResource,omitempty"`
	IconClicks       *IconClicks     `xml:"IconClicks,omitempty"`
	IconViewTracking []string        `xml:"IconViewTracking,omitempty"`
}

// StaticResource for icon
type StaticResource struct {
	CreativeType string `xml:"creativeType,attr"`
	URL          string `xml:",cdata"`
}

// IconClicks container
type IconClicks struct {
	IconClickThrough  string   `xml:"IconClickThrough,omitempty"`
	IconClickTracking []string `xml:"IconClickTracking,omitempty"`
}

// NonLinearAds container
type NonLinearAds struct {
	NonLinear []NonLinear `xml:"NonLinear"`
}

// NonLinear ad
type NonLinear struct {
	ID                     string          `xml:"id,attr,omitempty"`
	Width                  int             `xml:"width,attr"`
	Height                 int             `xml:"height,attr"`
	ExpandedWidth          int             `xml:"expandedWidth,attr,omitempty"`
	ExpandedHeight         int             `xml:"expandedHeight,attr,omitempty"`
	Scalable               bool            `xml:"scalable,attr,omitempty"`
	MaintainAspectRatio    bool            `xml:"maintainAspectRatio,attr,omitempty"`
	MinSuggestedDuration   string          `xml:"minSuggestedDuration,attr,omitempty"`
	APIFramework           string          `xml:"apiFramework,attr,omitempty"`
	StaticResource         *StaticResource `xml:"StaticResource,omitempty"`
	IFrameResource         string          `xml:"IFrameResource,omitempty"`
	HTMLResource           string          `xml:"HTMLResource,omitempty"`
	NonLinearClickThrough  string          `xml:"NonLinearClickThrough,omitempty"`
	NonLinearClickTracking []string        `xml:"NonLinearClickTracking,omitempty"`
	AdParameters           *AdParameters   `xml:"AdParameters,omitempty"`
}

// CompanionAds container
type CompanionAds struct {
	Required  string      `xml:"required,attr,omitempty"`
	Companion []Companion `xml:"Companion"`
}

// Companion ad
type Companion struct {
	ID                     string          `xml:"id,attr,omitempty"`
	Width                  int             `xml:"width,attr"`
	Height                 int             `xml:"height,attr"`
	AssetWidth             int             `xml:"assetWidth,attr,omitempty"`
	AssetHeight            int             `xml:"assetHeight,attr,omitempty"`
	ExpandedWidth          int             `xml:"expandedWidth,attr,omitempty"`
	ExpandedHeight         int             `xml:"expandedHeight,attr,omitempty"`
	APIFramework           string          `xml:"apiFramework,attr,omitempty"`
	AdSlotID               string          `xml:"adSlotId,attr,omitempty"`
	StaticResource         *StaticResource `xml:"StaticResource,omitempty"`
	IFrameResource         string          `xml:"IFrameResource,omitempty"`
	HTMLResource           string          `xml:"HTMLResource,omitempty"`
	AdParameters           *AdParameters   `xml:"AdParameters,omitempty"`
	AltText                string          `xml:"AltText,omitempty"`
	CompanionClickThrough  string          `xml:"CompanionClickThrough,omitempty"`
	CompanionClickTracking []string        `xml:"CompanionClickTracking,omitempty"`
	TrackingEvents         *TrackingEvents `xml:"TrackingEvents,omitempty"`
}

// Extensions container
type Extensions struct {
	Extension []Extension `xml:"Extension"`
}

// Extension element
type Extension struct {
	Type            string           `xml:"type,attr,omitempty"`
	AdVerifications *AdVerifications `xml:"AdVerifications,omitempty"`
	CustomTracking  *CustomTracking  `xml:"CustomTracking,omitempty"`
}

// AdVerifications for OMID
type AdVerifications struct {
	Verification []Verification `xml:"Verification"`
}

// Verification element
type Verification struct {
	Vendor                 string              `xml:"vendor,attr"`
	JavaScriptResource     *JavaScriptResource `xml:"JavaScriptResource,omitempty"`
	FlashResource          *FlashResource      `xml:"FlashResource,omitempty"`
	ViewableImpression     *ViewableImpression `xml:"ViewableImpression,omitempty"`
	VerificationParameters string              `xml:"VerificationParameters,omitempty"`
}

// JavaScriptResource for verification
type JavaScriptResource struct {
	APIFramework    string `xml:"apiFramework,attr"`
	BrowserOptional bool   `xml:"browserOptional,attr,omitempty"`
	URL             string `xml:",cdata"`
}

// FlashResource for verification
type FlashResource struct {
	APIFramework string `xml:"apiFramework,attr"`
	URL          string `xml:",cdata"`
}

// ViewableImpression for viewability
type ViewableImpression struct {
	ID               string   `xml:"id,attr,omitempty"`
	Viewable         []string `xml:"Viewable,omitempty"`
	NotViewable      []string `xml:"NotViewable,omitempty"`
	ViewUndetermined []string `xml:"ViewUndetermined,omitempty"`
}

// CustomTracking for custom events
type CustomTracking struct {
	Tracking []CustomTrackingEvent `xml:"Tracking"`
}

// CustomTrackingEvent element
type CustomTrackingEvent struct {
	Event string `xml:"event,attr"`
	URL   string `xml:",cdata"`
}

// OpenRTB types for integration

// OpenRTBRequest represents an OpenRTB 2.5/3.0 bid request
type OpenRTBRequest struct {
	ID      string              `json:"id"`
	Imp     []OpenRTBImpression `json:"imp"`
	Site    *Site               `json:"site,omitempty"`
	App     *App                `json:"app,omitempty"`
	Device  Device              `json:"device"`
	User    User                `json:"user"`
	Test    int                 `json:"test,omitempty"`
	AT      int                 `json:"at"`
	TMax    int                 `json:"tmax,omitempty"`
	WSeat   []string            `json:"wseat,omitempty"`
	BSeat   []string            `json:"bseat,omitempty"`
	AllImps int                 `json:"allimps,omitempty"`
	Cur     []string            `json:"cur,omitempty"`
	WLang   []string            `json:"wlang,omitempty"`
	BCat    []string            `json:"bcat,omitempty"`
	BAdv    []string            `json:"badv,omitempty"`
	BApp    []string            `json:"bapp,omitempty"`
	Source  Source              `json:"source,omitempty"`
	Regs    Regs                `json:"regs,omitempty"`
	Ext     interface{}         `json:"ext,omitempty"`
}

// OpenRTBResponse represents an OpenRTB bid response
type OpenRTBResponse struct {
	ID         string      `json:"id"`
	SeatBid    []SeatBid   `json:"seatbid,omitempty"`
	BidID      string      `json:"bidid,omitempty"`
	Cur        string      `json:"cur,omitempty"`
	CustomData string      `json:"customdata,omitempty"`
	NBR        int         `json:"nbr,omitempty"`
	Ext        interface{} `json:"ext,omitempty"`
}

// SeatBid represents a seat bid
type SeatBid struct {
	Bid   []Bid       `json:"bid"`
	Seat  string      `json:"seat,omitempty"`
	Group int         `json:"group,omitempty"`
	Ext   interface{} `json:"ext,omitempty"`
}

// Bid represents a bid
type Bid struct {
	ID             string      `json:"id"`
	ImpID          string      `json:"impid"`
	Price          float64     `json:"price"`
	AdID           string      `json:"adid,omitempty"`
	NURL           string      `json:"nurl,omitempty"`
	BURL           string      `json:"burl,omitempty"`
	LURL           string      `json:"lurl,omitempty"`
	ADM            string      `json:"adm,omitempty"`
	ADURL          string      `json:"adurl,omitempty"`
	ADomain        []string    `json:"adomain,omitempty"`
	Bundle         string      `json:"bundle,omitempty"`
	IURL           string      `json:"iurl,omitempty"`
	CID            string      `json:"cid,omitempty"`
	CrID           string      `json:"crid,omitempty"`
	Tactic         string      `json:"tactic,omitempty"`
	Cat            []string    `json:"cat,omitempty"`
	Attr           []int       `json:"attr,omitempty"`
	API            int         `json:"api,omitempty"`
	Protocol       int         `json:"protocol,omitempty"`
	QAGMediaRating int         `json:"qagmediarating,omitempty"`
	Language       string      `json:"language,omitempty"`
	DealID         string      `json:"dealid,omitempty"`
	W              int         `json:"w,omitempty"`
	H              int         `json:"h,omitempty"`
	WRatio         int         `json:"wratio,omitempty"`
	HRatio         int         `json:"hratio,omitempty"`
	Exp            int         `json:"exp,omitempty"`
	Cur            string      `json:"cur,omitempty"`
	Ext            interface{} `json:"ext,omitempty"`
}

// Video object for video impressions
type Video struct {
	MIMEs          []string `json:"mimes"`
	MinDuration    int      `json:"minduration,omitempty"`
	MaxDuration    int      `json:"maxduration,omitempty"`
	Protocols      []int    `json:"protocols,omitempty"`
	W              int      `json:"w,omitempty"`
	H              int      `json:"h,omitempty"`
	StartDelay     *int     `json:"startdelay,omitempty"`
	Placement      int      `json:"placement,omitempty"`
	Linearity      int      `json:"linearity,omitempty"`
	Skip           int      `json:"skip,omitempty"`
	SkipMin        int      `json:"skipmin,omitempty"`
	SkipAfter      int      `json:"skipafter,omitempty"`
	Sequence       int      `json:"sequence,omitempty"`
	BAttr          []int    `json:"battr,omitempty"`
	MaxExtended    int      `json:"maxextended,omitempty"`
	MinBitrate     int      `json:"minbitrate,omitempty"`
	MaxBitrate     int      `json:"maxbitrate,omitempty"`
	BoxingAllowed  int      `json:"boxingallowed,omitempty"`
	PlaybackMethod []int    `json:"playbackmethod,omitempty"`
	PlaybackEnd    int      `json:"playbackend,omitempty"`
	Delivery       []int    `json:"delivery,omitempty"`
	Pos            int      `json:"pos,omitempty"`
	CompanionAd    []Banner `json:"companionad,omitempty"`
	API            []int    `json:"api,omitempty"`
	CompanionType  []int    `json:"companiontype,omitempty"`
	// CTV/OTT specific
	PodID       string      `json:"podid,omitempty"`
	PodSequence int         `json:"podseq,omitempty"`
	RqdDurs     []int       `json:"rqddurs,omitempty"`
	MaxSeq      int         `json:"maxseq,omitempty"`
	MinAds      int         `json:"minads,omitempty"`
	MaxAds      int         `json:"maxads,omitempty"`
	Ext         interface{} `json:"ext,omitempty"`
}

// Banner object
type Banner struct {
	W        int         `json:"w,omitempty"`
	H        int         `json:"h,omitempty"`
	WMax     int         `json:"wmax,omitempty"`
	HMax     int         `json:"hmax,omitempty"`
	WMin     int         `json:"wmin,omitempty"`
	HMin     int         `json:"hmin,omitempty"`
	BType    []int       `json:"btype,omitempty"`
	BAttr    []int       `json:"battr,omitempty"`
	Pos      int         `json:"pos,omitempty"`
	MIMEs    []string    `json:"mimes,omitempty"`
	TopFrame int         `json:"topframe,omitempty"`
	ExpDir   []int       `json:"expdir,omitempty"`
	API      []int       `json:"api,omitempty"`
	ID       string      `json:"id,omitempty"`
	VCM      int         `json:"vcm,omitempty"`
	Ext      interface{} `json:"ext,omitempty"`
}

// Device object
type Device struct {
	UA             string      `json:"ua,omitempty"`
	Geo            Geo         `json:"geo,omitempty"`
	DNT            int         `json:"dnt,omitempty"`
	LMT            int         `json:"lmt,omitempty"`
	IP             string      `json:"ip,omitempty"`
	IPv6           string      `json:"ipv6,omitempty"`
	DeviceType     int         `json:"devicetype,omitempty"`
	Make           string      `json:"make,omitempty"`
	Model          string      `json:"model,omitempty"`
	OS             string      `json:"os,omitempty"`
	OSV            string      `json:"osv,omitempty"`
	HWV            string      `json:"hwv,omitempty"`
	H              int         `json:"h,omitempty"`
	W              int         `json:"w,omitempty"`
	PPI            int         `json:"ppi,omitempty"`
	PxRatio        float64     `json:"pxratio,omitempty"`
	JS             int         `json:"js,omitempty"`
	GeoFetch       int         `json:"geofetch,omitempty"`
	FlashVer       string      `json:"flashver,omitempty"`
	Language       string      `json:"language,omitempty"`
	Carrier        string      `json:"carrier,omitempty"`
	MCCMNC         string      `json:"mccmnc,omitempty"`
	ConnectionType int         `json:"connectiontype,omitempty"`
	IFA            string      `json:"ifa,omitempty"`
	DPIDSHA1       string      `json:"dpidsha1,omitempty"`
	DPIDMD5        string      `json:"dpidmd5,omitempty"`
	MACSHA1        string      `json:"macsha1,omitempty"`
	MACMD5         string      `json:"macmd5,omitempty"`
	Ext            interface{} `json:"ext,omitempty"`
}

// Geo object
type Geo struct {
	Lat           float64     `json:"lat,omitempty"`
	Lon           float64     `json:"lon,omitempty"`
	Type          int         `json:"type,omitempty"`
	Accuracy      int         `json:"accuracy,omitempty"`
	LastFix       int         `json:"lastfix,omitempty"`
	IPService     int         `json:"ipservice,omitempty"`
	Country       string      `json:"country,omitempty"`
	Region        string      `json:"region,omitempty"`
	RegionFIPS104 string      `json:"regionfips104,omitempty"`
	Metro         string      `json:"metro,omitempty"`
	City          string      `json:"city,omitempty"`
	ZIP           string      `json:"zip,omitempty"`
	UTCOffset     int         `json:"utcoffset,omitempty"`
	Ext           interface{} `json:"ext,omitempty"`
}

// User object
type User struct {
	ID         string      `json:"id,omitempty"`
	BuyerUID   string      `json:"buyeruid,omitempty"`
	YOB        int         `json:"yob,omitempty"`
	Gender     string      `json:"gender,omitempty"`
	Keywords   string      `json:"keywords,omitempty"`
	CustomData string      `json:"customdata,omitempty"`
	Geo        *Geo        `json:"geo,omitempty"`
	Data       []Data      `json:"data,omitempty"`
	Consent    string      `json:"consent,omitempty"`
	Ext        interface{} `json:"ext,omitempty"`
}

// Data object
type Data struct {
	ID      string      `json:"id,omitempty"`
	Name    string      `json:"name,omitempty"`
	Segment []Segment   `json:"segment,omitempty"`
	Ext     interface{} `json:"ext,omitempty"`
}

// Segment object
type Segment struct {
	ID    string      `json:"id,omitempty"`
	Name  string      `json:"name,omitempty"`
	Value string      `json:"value,omitempty"`
	Ext   interface{} `json:"ext,omitempty"`
}

// App object
type App struct {
	ID            string      `json:"id,omitempty"`
	Name          string      `json:"name,omitempty"`
	Bundle        string      `json:"bundle,omitempty"`
	Domain        string      `json:"domain,omitempty"`
	StoreURL      string      `json:"storeurl,omitempty"`
	Cat           []string    `json:"cat,omitempty"`
	SectionCat    []string    `json:"sectioncat,omitempty"`
	PageCat       []string    `json:"pagecat,omitempty"`
	Ver           string      `json:"ver,omitempty"`
	PrivacyPolicy int         `json:"privacypolicy,omitempty"`
	Paid          int         `json:"paid,omitempty"`
	Publisher     *Publisher  `json:"publisher,omitempty"`
	Content       Content     `json:"content,omitempty"`
	Keywords      string      `json:"keywords,omitempty"`
	Ext           interface{} `json:"ext,omitempty"`
}

// Site object
type Site struct {
	ID            string      `json:"id,omitempty"`
	Name          string      `json:"name,omitempty"`
	Domain        string      `json:"domain,omitempty"`
	Cat           []string    `json:"cat,omitempty"`
	SectionCat    []string    `json:"sectioncat,omitempty"`
	PageCat       []string    `json:"pagecat,omitempty"`
	Page          string      `json:"page,omitempty"`
	Ref           string      `json:"ref,omitempty"`
	Search        string      `json:"search,omitempty"`
	Mobile        int         `json:"mobile,omitempty"`
	PrivacyPolicy int         `json:"privacypolicy,omitempty"`
	Publisher     *Publisher  `json:"publisher,omitempty"`
	Content       *Content    `json:"content,omitempty"`
	Keywords      string      `json:"keywords,omitempty"`
	Ext           interface{} `json:"ext,omitempty"`
}

// Publisher object
type Publisher struct {
	ID     string      `json:"id,omitempty"`
	Name   string      `json:"name,omitempty"`
	Cat    []string    `json:"cat,omitempty"`
	Domain string      `json:"domain,omitempty"`
	Ext    interface{} `json:"ext,omitempty"`
}

// Content object
type Content struct {
	ID                 string      `json:"id,omitempty"`
	Episode            string      `json:"episode,omitempty"`
	Title              string      `json:"title,omitempty"`
	Series             string      `json:"series,omitempty"`
	Season             string      `json:"season,omitempty"`
	Artist             string      `json:"artist,omitempty"`
	Genre              string      `json:"genre,omitempty"`
	Album              string      `json:"album,omitempty"`
	ISRC               string      `json:"isrc,omitempty"`
	Producer           *Producer   `json:"producer,omitempty"`
	URL                string      `json:"url,omitempty"`
	Cat                []string    `json:"cat,omitempty"`
	ProdQ              int         `json:"prodq,omitempty"`
	VideoQuality       int         `json:"videoquality,omitempty"`
	Context            int         `json:"context,omitempty"`
	ContentRating      string      `json:"contentrating,omitempty"`
	Rating             string      `json:"rating,omitempty"`
	UserRating         string      `json:"userrating,omitempty"`
	QAGMediaRating     int         `json:"qagmediarating,omitempty"`
	Keywords           string      `json:"keywords,omitempty"`
	LiveStream         int         `json:"livestream,omitempty"`
	SourceRelationship int         `json:"sourcerelationship,omitempty"`
	Len                int         `json:"len,omitempty"`
	Language           string      `json:"language,omitempty"`
	Embeddable         int         `json:"embeddable,omitempty"`
	Data               []Data      `json:"data,omitempty"`
	Network            *Network    `json:"network,omitempty"`
	Channel            *Channel    `json:"channel,omitempty"`
	Ext                interface{} `json:"ext,omitempty"`
}

// Producer object
type Producer struct {
	ID     string      `json:"id,omitempty"`
	Name   string      `json:"name,omitempty"`
	Cat    []string    `json:"cat,omitempty"`
	Domain string      `json:"domain,omitempty"`
	Ext    interface{} `json:"ext,omitempty"`
}

// Network object
type Network struct {
	ID     string      `json:"id,omitempty"`
	Name   string      `json:"name,omitempty"`
	Domain string      `json:"domain,omitempty"`
	Ext    interface{} `json:"ext,omitempty"`
}

// Channel object
type Channel struct {
	ID     string      `json:"id,omitempty"`
	Name   string      `json:"name,omitempty"`
	Domain string      `json:"domain,omitempty"`
	Ext    interface{} `json:"ext,omitempty"`
}

// Source object
type Source struct {
	FD     int          `json:"fd,omitempty"`
	TID    string       `json:"tid,omitempty"`
	PChain string       `json:"pchain,omitempty"`
	SKAdN  *SKAdNetwork `json:"skadn,omitempty"`
	Ext    interface{}  `json:"ext,omitempty"`
}

// SKAdNetwork object
type SKAdNetwork struct {
	Version    string      `json:"version,omitempty"`
	SourceApp  string      `json:"sourceapp,omitempty"`
	SKAdNetIDs []string    `json:"skadnetids,omitempty"`
	Ext        interface{} `json:"ext,omitempty"`
}

// Regs object (Regulations)
type Regs struct {
	COPPA int         `json:"coppa,omitempty"`
	GDPR  int         `json:"gdpr,omitempty"`
	CCPA  string      `json:"us_privacy,omitempty"`
	Ext   interface{} `json:"ext,omitempty"`
}

// Metric object
type Metric struct {
	Type   string      `json:"type"`
	Value  float64     `json:"value"`
	Vendor string      `json:"vendor,omitempty"`
	Ext    interface{} `json:"ext,omitempty"`
}

// Audio object
type Audio struct {
	MIMEs         []string    `json:"mimes"`
	MinDuration   int         `json:"minduration,omitempty"`
	MaxDuration   int         `json:"maxduration,omitempty"`
	Protocols     []int       `json:"protocols,omitempty"`
	StartDelay    int         `json:"startdelay,omitempty"`
	Sequence      int         `json:"sequence,omitempty"`
	BAttr         []int       `json:"battr,omitempty"`
	MaxExtended   int         `json:"maxextended,omitempty"`
	MinBitrate    int         `json:"minbitrate,omitempty"`
	MaxBitrate    int         `json:"maxbitrate,omitempty"`
	Delivery      []int       `json:"delivery,omitempty"`
	CompanionAd   []Banner    `json:"companionad,omitempty"`
	API           []int       `json:"api,omitempty"`
	CompanionType []int       `json:"companiontype,omitempty"`
	MaxSeq        int         `json:"maxseq,omitempty"`
	Feed          int         `json:"feed,omitempty"`
	Stitched      int         `json:"stitched,omitempty"`
	NVol          int         `json:"nvol,omitempty"`
	Ext           interface{} `json:"ext,omitempty"`
}

// Native object
type Native struct {
	Request string      `json:"request"`
	Ver     string      `json:"ver,omitempty"`
	API     []int       `json:"api,omitempty"`
	BAttr   []int       `json:"battr,omitempty"`
	Ext     interface{} `json:"ext,omitempty"`
}

// PMP (Private Marketplace) object
type PMP struct {
	PrivateAuction int         `json:"private_auction,omitempty"`
	Deals          []Deal      `json:"deals,omitempty"`
	Ext            interface{} `json:"ext,omitempty"`
}

// Deal object
type Deal struct {
	ID          string      `json:"id"`
	BidFloor    float64     `json:"bidfloor,omitempty"`
	BidFloorCur string      `json:"bidfloorcur,omitempty"`
	AT          int         `json:"at,omitempty"`
	WSeat       []string    `json:"wseat,omitempty"`
	WAdvDomains []string    `json:"wadomain,omitempty"`
	Ext         interface{} `json:"ext,omitempty"`
}

// OpenRTBImpression for bid requests
type OpenRTBImpression struct {
	ID                string      `json:"id"`
	Metric            []Metric    `json:"metric,omitempty"`
	Banner            *Banner     `json:"banner,omitempty"`
	Video             *Video      `json:"video,omitempty"`
	Audio             *Audio      `json:"audio,omitempty"`
	Native            *Native     `json:"native,omitempty"`
	PMP               *PMP        `json:"pmp,omitempty"`
	DisplayManager    string      `json:"displaymanager,omitempty"`
	DisplayManagerVer string      `json:"displaymanagerver,omitempty"`
	Instl             int         `json:"instl,omitempty"`
	TagID             string      `json:"tagid,omitempty"`
	BidFloor          float64     `json:"bidfloor,omitempty"`
	BidFloorCur       string      `json:"bidfloorcur,omitempty"`
	ClickBrowser      int         `json:"clickbrowser,omitempty"`
	Secure            int         `json:"secure,omitempty"`
	IframeBuster      []string    `json:"iframebuster,omitempty"`
	Exp               int         `json:"exp,omitempty"`
	Ext               interface{} `json:"ext,omitempty"`
}

// Storage and Analytics types

// ImpressionRecord for tracking
type ImpressionRecord struct {
	ID        string       `json:"id"`
	Timestamp time.Time    `json:"timestamp"`
	AppToken  string       `json:"app_token"`
	ZoneID    int          `json:"zone_id"`
	Device    DeviceInfo   `json:"device"`
	Location  LocationInfo `json:"location"`
	AdCount   int          `json:"ad_count"`
	Revenue   float64      `json:"revenue"`
}

// DeviceInfo for tracking
type DeviceInfo struct {
	OS        string `json:"os"`
	OSVersion string `json:"os_version"`
	Model     string `json:"model"`
	IFA       string `json:"ifa"`
	Type      string `json:"type"`
}

// LocationInfo for tracking
type LocationInfo struct {
	Lat     string `json:"lat"`
	Lon     string `json:"lon"`
	Country string `json:"country"`
	Region  string `json:"region"`
	City    string `json:"city"`
}

// Manager interfaces

// RTBExchange interface
type RTBExchange interface {
	RunAuction(ctx context.Context, req *OpenRTBRequest) (*OpenRTBResponse, error)
}

// StorageBackend interface
type StorageBackend interface {
	StoreImpression(imp *ImpressionRecord) error
	GetImpression(id string) (*ImpressionRecord, error)
}

// AnalyticsEngine interface
type AnalyticsEngine interface {
	TrackImpression(imp *ImpressionRecord)
	TrackClick(clickID string, impID string)
	GetMetrics(startTime, endTime time.Time) map[string]interface{}
}

// PrivacyManager interface
type PrivacyManager interface {
	CheckCompliance(consent string, gdpr int, ccpa string) bool
	AnonymizeData(data interface{}) interface{}
}

// BlockchainManager interface
type BlockchainManager interface {
	RecordImpression(imp *ImpressionRecord, wallet string, chainID int) error
	VerifyProofOfView(pov string) bool
	ProcessPayment(wallet string, amount float64, chainID int) error
}

// RateLimiter for DSP connections
type RateLimiter struct {
	QPS      int
	Requests chan time.Time
}

// NewRateLimiter creates a rate limiter
func NewRateLimiter(qps int) *RateLimiter {
	rl := &RateLimiter{
		QPS:      qps,
		Requests: make(chan time.Time, qps),
	}

	// Start ticker to refill tokens
	go func() {
		ticker := time.NewTicker(time.Second / time.Duration(qps))
		for t := range ticker.C {
			select {
			case rl.Requests <- t:
			default:
			}
		}
	}()

	return rl
}

// Allow checks if request is allowed
func (rl *RateLimiter) Allow() bool {
	select {
	case <-rl.Requests:
		return true
	default:
		return false
	}
}

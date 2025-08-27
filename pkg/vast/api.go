package vast

import (
	"encoding/xml"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// VASTRequest represents the complete VAST API request parameters
// Compatible with PubNative and other major ad servers
type VASTRequest struct {
	// Required Core Parameters
	AppToken    string `form:"apptoken" binding:"required" json:"apptoken"`
	OS          string `form:"os" binding:"required,oneof=ios android ctv" json:"os"`
	OSVer       string `form:"osver" binding:"required" json:"osver"`
	DeviceModel string `form:"devicemodel" binding:"required" json:"devicemodel"`
	DNT         int    `form:"dnt" binding:"required,min=0,max=1" json:"dnt"`
	AL          string `form:"al" binding:"required,oneof=s m l xl" json:"al"` // Ad layout: s=small, m=medium, l=large, xl=extra large
	ZoneID      int    `form:"zoneid" binding:"required" json:"zoneid"`

	// Device Identifiers (at least one required)
	IDFA     string `form:"idfa" json:"idfa"`         // iOS Advertising ID
	GID      string `form:"gid" json:"gid"`           // Android Advertising ID
	IFV      string `form:"ifv" json:"ifv"`           // iOS Identifier for Vendor
	UID      string `form:"uid" json:"uid"`           // User identifier (when IDFA/GID not available)
	IDFAMD5  string `form:"idfamd5" json:"idfamd5"`   // IDFA in MD5
	IDFASHA1 string `form:"idfasha1" json:"idfasha1"` // IDFA in SHA1
	GIDMD5   string `form:"gidmd5" json:"gidmd5"`     // GID in MD5
	GIDSHA1  string `form:"gidsha1" json:"gidsha1"`   // GID in SHA1

	// Server-to-Server Parameters
	SRVI int    `form:"srvi" json:"srvi"` // 1 if server-side request, 0 if client-side
	UA   string `form:"ua" json:"ua"`     // User Agent (required when srvi=1)
	IP   string `form:"ip" json:"ip"`     // IP address (required when srvi=1)
	IPV6 string `form:"ipv6" json:"ipv6"` // IPv6 address

	// Targeting Parameters
	AdCount  int    `form:"adcount" default:"1" json:"adcount"`         // Number of ads to return (1-20)
	Locale   string `form:"locale" json:"locale"`                       // Device locale (e.g., en_US)
	Lat      string `form:"lat" json:"lat"`                             // Latitude
	Long     string `form:"long" json:"long"`                           // Longitude
	Gender   string `form:"gender" json:"gender"`                       // m, f, o (other)
	Age      int    `form:"age" json:"age"`                             // User age
	Keywords string `form:"keywords" json:"keywords"`                   // Comma-separated keywords
	BundleID string `form:"bundleid" json:"bundleid"`                   // App bundle ID
	AppVer   string `form:"appver" json:"appver"`                       // App version
	Secure   int    `form:"secure" default:"1" json:"secure"`          // HTTPS (1) or HTTP (0)

	// Privacy Compliance
	COPPA        int    `form:"coppa" json:"coppa"`               // COPPA flag (0=NO, 1=YES)
	USPrivacy    string `form:"usprivacy" json:"usprivacy"`       // CCPA string (e.g., "1YNN")
	UserConsent  string `form:"userconsent" json:"userconsent"`   // GDPR consent string
	GDPR         int    `form:"gdpr" json:"gdpr"`                 // GDPR applies (0=NO, 1=YES)

	// Video/CTV Specific
	RV              string `form:"rv" json:"rv"`                           // Rewarded video flag
	DH              string `form:"dh" json:"dh"`                           // Device height in pixels
	DW              string `form:"dw" json:"dw"`                           // Device width in pixels
	MinVideoDur     int    `form:"mindur" json:"mindur"`                   // Min video duration
	MaxVideoDur     int    `form:"maxdur" json:"maxdur"`                   // Max video duration
	StartDelay      int    `form:"startdelay" json:"startdelay"`           // Video start delay
	Linearity       int    `form:"linearity" json:"linearity"`             // Linear=1, NonLinear=2
	Skip            int    `form:"skip" json:"skip"`                       // Skippable (0=NO, 1=YES)
	SkipMin         int    `form:"skipmin" json:"skipmin"`                 // Skip button delay
	SkipAfter       int    `form:"skipafter" json:"skipafter"`             // Force skip after seconds
	Playbackmethod  []int  `form:"playbackmethod" json:"playbackmethod"`   // Playback methods
	PlayerSize      string `form:"playersize" json:"playersize"`           // WxH format

	// OMID (Open Measurement)
	OMIDPN string `form:"omidpn" json:"omidpn"` // OMID Partner name
	OMIDPV string `form:"omidpv" json:"omidpv"` // OMID Partner version

	// SKAdNetwork (iOS Attribution)
	SKAdNVersion   string   `form:"skadn_version" json:"skadn_version"`     // SKAdNetwork version (2.0+)
	SKAdNSourceApp string   `form:"skadn_sourceapp" json:"skadn_sourceapp"` // Publisher app ID
	SKAdNetIDs     []string `form:"skadnetids" json:"skadnetids"`           // DSP-specific SKAdNetwork IDs

	// Contextual Data
	InputLanguage    string `form:"inputlanguage" json:"inputlanguage"`       // Keyboard languages
	Battery          string `form:"battery" json:"battery"`                   // Battery percentage
	SessionDuration  string `form:"sessionduration" json:"sessionduration"`   // Session time in seconds
	AgeRating        string `form:"agerating" json:"agerating"`               // Content rating
	PubDomain        string `form:"pub_domain" json:"pub_domain"`             // Publisher domain

	// CTV/OTT Specific
	ContentID        string `form:"contentid" json:"contentid"`               // Content ID
	ContentTitle     string `form:"contenttitle" json:"contenttitle"`         // Content title
	ContentSeries    string `form:"contentseries" json:"contentseries"`       // Series name
	ContentSeason    string `form:"contentseason" json:"contentseason"`       // Season number
	ContentEpisode   string `form:"contentepisode" json:"contentepisode"`     // Episode number
	ContentGenre     string `form:"contentgenre" json:"contentgenre"`         // Content genre
	ContentRating    string `form:"contentrating" json:"contentrating"`       // Content rating
	ContentLength    int    `form:"contentlen" json:"contentlen"`             // Content length in seconds
	ContentLiveStream int    `form:"livestream" json:"livestream"`             // Live stream flag
	ContentLanguage  string `form:"contentlang" json:"contentlang"`           // Content language

	// Ad Pod Support (CTV)
	PodID            string `form:"podid" json:"podid"`                       // Ad pod ID
	PodSequence      int    `form:"podseq" json:"podseq"`                     // Position in pod
	TotalPods        int    `form:"totalpods" json:"totalpods"`               // Total pods in content
	PodMaxDuration   int    `form:"podmaxdur" json:"podmaxdur"`               // Max pod duration
	PodMinAds        int    `form:"podminads" json:"podminads"`               // Min ads per pod
	PodMaxAds        int    `form:"podmaxads" json:"podmaxads"`               // Max ads per pod

	// Blockchain Specific (Lux ADX Extensions)
	WalletAddress    string `form:"wallet" json:"wallet"`                     // User wallet for rewards
	ChainID          int    `form:"chainid" json:"chainid"`                   // Blockchain ID
	SmartContract    string `form:"contract" json:"contract"`                 // Smart contract address
	OnChainTracking  int    `form:"onchain" json:"onchain"`                   // On-chain tracking (0=NO, 1=YES)
	DecentralizedID  string `form:"did" json:"did"`                           // Decentralized ID
	ProofOfView      string `form:"pov" json:"pov"`                           // Proof of view hash
}

// VASTHandler handles VAST API requests with full parameter support
type VASTHandler struct {
	Exchange      *RTBExchange
	Storage       StorageBackend
	Analytics     *AnalyticsEngine
	PrivacyMgr    *PrivacyManager
	BlockchainMgr *BlockchainManager
}

// HandleVASTRequest processes VAST API requests
func (h *VASTHandler) HandleVASTRequest(c *gin.Context) {
	var req VASTRequest
	
	// Bind query parameters
	if err := c.ShouldBindQuery(&req); err != nil {
		c.XML(http.StatusBadRequest, VASTError{
			Code:    400,
			Message: "Invalid request parameters: " + err.Error(),
		})
		return
	}

	// Validate server-to-server requirements
	if req.SRVI == 1 && (req.UA == "" || req.IP == "") {
		c.XML(http.StatusBadRequest, VASTError{
			Code:    400,
			Message: "Server-side requests require 'ua' and 'ip' parameters",
		})
		return
	}

	// Privacy compliance checks
	if err := h.checkPrivacyCompliance(&req); err != nil {
		c.XML(http.StatusNoContent, nil) // No ads due to privacy
		return
	}

	// Build OpenRTB request from VAST parameters
	rtbReq := h.buildOpenRTBRequest(&req)

	// Run auction
	rtbResp, err := h.Exchange.RunAuction(c.Request.Context(), rtbReq)
	if err != nil || len(rtbResp.SeatBid) == 0 {
		c.XML(http.StatusNoContent, nil) // No ads available
		return
	}

	// Convert OpenRTB response to VAST
	vast := h.buildVASTResponse(&req, rtbResp)

	// Track impression (async)
	go h.trackImpression(&req, vast)

	// Set cache headers for CDN
	c.Header("Cache-Control", "private, max-age=300")
	c.Header("X-ADX-Request-ID", rtbReq.ID)

	// Return VAST XML
	c.XML(http.StatusOK, vast)
}

// buildOpenRTBRequest converts VAST request to OpenRTB
func (h *VASTHandler) buildOpenRTBRequest(req *VASTRequest) *OpenRTBRequest {
	rtb := &OpenRTBRequest{
		ID:     uuid.New().String(),
		Imp:    []Impression{},
		Device: Device{},
		User:   User{},
		App:    App{},
		Regs:   Regs{},
		Source: Source{},
	}

	// Build impression
	imp := Impression{
		ID:       "1",
		Video:    &Video{},
		Secure:   req.Secure,
		BidFloor: 0.01, // Default floor
		BidFloorCur: "USD",
	}

	// Video parameters
	if req.MaxVideoDur > 0 {
		imp.Video.MaxDuration = req.MaxVideoDur
	}
	if req.MinVideoDur > 0 {
		imp.Video.MinDuration = req.MinVideoDur
	}
	imp.Video.StartDelay = &req.StartDelay
	imp.Video.Linearity = req.Linearity
	imp.Video.Skip = req.Skip
	imp.Video.SkipMin = req.SkipMin
	imp.Video.SkipAfter = req.SkipAfter
	imp.Video.PlaybackMethod = req.Playbackmethod

	// Parse player size
	if req.PlayerSize != "" {
		parts := strings.Split(req.PlayerSize, "x")
		if len(parts) == 2 {
			if w, err := strconv.Atoi(parts[0]); err == nil {
				imp.Video.W = w
			}
			if h, err := strconv.Atoi(parts[1]); err == nil {
				imp.Video.H = h
			}
		}
	}

	// Ad pod support for CTV
	if req.PodID != "" {
		imp.Video.PodID = req.PodID
		imp.Video.PodSequence = req.PodSequence
		imp.Video.MaxSeq = req.TotalPods
		imp.Video.MaxExtended = req.PodMaxDuration
		imp.Video.MinAds = req.PodMinAds
		imp.Video.MaxAds = req.PodMaxAds
	}

	// VAST protocol versions
	imp.Video.Protocols = []int{2, 3, 5, 6, 7, 8} // VAST 2.0, 3.0, 4.0, 4.1, 4.2, 4.3
	imp.Video.MIMEs = []string{"video/mp4", "video/webm", "application/x-mpegURL"}

	// Set impression based on ad layout
	switch req.AL {
	case "s":
		imp.Video.W, imp.Video.H = 320, 180
	case "m":
		imp.Video.W, imp.Video.H = 640, 360
	case "l":
		imp.Video.W, imp.Video.H = 1280, 720
	case "xl":
		imp.Video.W, imp.Video.H = 1920, 1080
	}

	// Add impression
	for i := 0; i < req.AdCount && i < 20; i++ {
		impCopy := imp
		impCopy.ID = strconv.Itoa(i + 1)
		rtb.Imp = append(rtb.Imp, impCopy)
	}

	// Device information
	rtb.Device = Device{
		UA:          req.UA,
		IP:          req.IP,
		IPv6:        req.IPV6,
		DeviceType:  h.getDeviceType(req.DeviceModel),
		Make:        h.getDeviceMake(req.DeviceModel),
		Model:       req.DeviceModel,
		OS:          req.OS,
		OSV:         req.OSVer,
		Language:    req.Locale,
		IFA:         h.getIFA(req),
		DNT:         req.DNT,
		LMT:         req.DNT,
		Geo:         Geo{},
	}

	// Handle device dimensions
	if req.DW != "" {
		if w, err := strconv.Atoi(req.DW); err == nil {
			rtb.Device.W = w
		}
	}
	if req.DH != "" {
		if h, err := strconv.Atoi(req.DH); err == nil {
			rtb.Device.H = h
		}
	}

	// Geolocation
	if req.Lat != "" && req.Long != "" {
		if lat, err := strconv.ParseFloat(req.Lat, 64); err == nil {
			rtb.Device.Geo.Lat = lat
		}
		if lon, err := strconv.ParseFloat(req.Long, 64); err == nil {
			rtb.Device.Geo.Lon = lon
		}
	}

	// User information
	rtb.User = User{
		ID:       req.UID,
		Gender:   req.Gender,
		Keywords: req.Keywords,
	}
	if req.Age > 0 {
		rtb.User.YOB = time.Now().Year() - req.Age
	}

	// App information
	rtb.App = App{
		ID:       req.AppToken,
		Bundle:   req.BundleID,
		Ver:      req.AppVer,
		Domain:   req.PubDomain,
		Keywords: req.Keywords,
		Content:  Content{},
	}

	// Content information for CTV
	if req.ContentID != "" {
		rtb.App.Content = Content{
			ID:       req.ContentID,
			Title:    req.ContentTitle,
			Series:   req.ContentSeries,
			Season:   req.ContentSeason,
			Episode:  req.ContentEpisode,
			Genre:    req.ContentGenre,
			Rating:   req.ContentRating,
			Len:      req.ContentLength,
			LiveStream: req.ContentLiveStream,
			Language: req.ContentLanguage,
		}
	}

	// Privacy regulations
	rtb.Regs = Regs{
		COPPA: req.COPPA,
		GDPR:  req.GDPR,
		CCPA:  req.USPrivacy,
	}

	// SKAdNetwork for iOS
	if req.SKAdNVersion != "" {
		rtb.Source.SKAdN = &SKAdNetwork{
			Version:    req.SKAdNVersion,
			SourceApp:  req.SKAdNSourceApp,
			SKAdNetIDs: req.SKAdNetIDs,
		}
	}

	// Blockchain extensions
	if req.WalletAddress != "" {
		rtb.Ext = map[string]interface{}{
			"blockchain": map[string]interface{}{
				"wallet":    req.WalletAddress,
				"chainid":   req.ChainID,
				"contract":  req.SmartContract,
				"onchain":   req.OnChainTracking == 1,
				"did":       req.DecentralizedID,
				"pov":       req.ProofOfView,
			},
		}
	}

	return rtb
}

// buildVASTResponse creates VAST XML from OpenRTB response
func (h *VASTHandler) buildVASTResponse(req *VASTRequest, rtbResp *OpenRTBResponse) *VAST {
	vast := &VAST{
		Version: "4.3",
		Ads:     []Ad{},
	}

	for _, seatBid := range rtbResp.SeatBid {
		for _, bid := range seatBid.Bid {
			ad := h.createVASTAd(req, &bid)
			vast.Ads = append(vast.Ads, ad)
		}
	}

	return vast
}

// createVASTAd creates a VAST Ad from OpenRTB Bid
func (h *VASTHandler) createVASTAd(req *VASTRequest, bid *Bid) Ad {
	ad := Ad{
		ID: bid.ID,
		InLine: &InLine{
			AdSystem: AdSystem{
				Name:    "Lux ADX",
				Version: "1.0",
			},
			AdTitle:     bid.ADomain[0] + " Video Ad",
			Description: "Video advertisement",
			Advertiser:  bid.ADomain[0],
			Pricing: &Pricing{
				Model:    "CPM",
				Currency: bid.Cur,
				Value:    bid.Price,
			},
			Impression: []Impression{},
			Creatives:  Creatives{},
		},
	}

	// Add impression tracking
	ad.InLine.Impression = append(ad.InLine.Impression, Impression{
		ID:  "main",
		URL: h.buildTrackingURL("impression", req, bid),
	})

	// Add error tracking
	ad.InLine.Error = append(ad.InLine.Error, h.buildTrackingURL("error", req, bid))

	// Create video creative
	creative := Creative{
		ID:   "1",
		AdID: bid.ID,
		Linear: &Linear{
			Duration: formatDuration(30), // Default 30s
			MediaFiles: MediaFiles{
				MediaFile: []MediaFile{},
			},
			VideoClicks: &VideoClicks{
				ClickThrough: &ClickThrough{
					URL: bid.NURL, // Click URL
				},
				ClickTracking: []ClickTracking{
					{URL: h.buildTrackingURL("click", req, bid)},
				},
			},
			TrackingEvents: &TrackingEvents{
				Tracking: []Tracking{
					{Event: "start", URL: h.buildTrackingURL("start", req, bid)},
					{Event: "firstQuartile", URL: h.buildTrackingURL("firstQuartile", req, bid)},
					{Event: "midpoint", URL: h.buildTrackingURL("midpoint", req, bid)},
					{Event: "thirdQuartile", URL: h.buildTrackingURL("thirdQuartile", req, bid)},
					{Event: "complete", URL: h.buildTrackingURL("complete", req, bid)},
				},
			},
		},
	}

	// Add media files based on ad layout
	mediaFiles := h.getMediaFilesForLayout(req.AL, bid.ADURL)
	creative.Linear.MediaFiles.MediaFile = mediaFiles

	// Add skip offset if applicable
	if req.Skip == 1 && req.SkipMin > 0 {
		creative.Linear.SkipOffset = fmt.Sprintf("00:00:%02d", req.SkipMin)
	}

	ad.InLine.Creatives.Creative = append(ad.InLine.Creatives.Creative, creative)

	// Add OMID verification if present
	if req.OMIDPN != "" {
		ad.InLine.Extensions = &Extensions{
			Extension: []Extension{
				{
					Type: "AdVerifications",
					AdVerifications: &AdVerifications{
						Verification: []Verification{
							{
								Vendor: req.OMIDPN,
								JavaScriptResource: &JavaScriptResource{
									APIFramework: "omid",
									URL:          h.getOMIDVerificationScript(req.OMIDPN),
								},
								VerificationParameters: fmt.Sprintf(`{"partnername":"%s","partnerversion":"%s"}`, req.OMIDPN, req.OMIDPV),
							},
						},
					},
				},
			},
		}
	}

	return ad
}

// Helper functions

func (h *VASTHandler) checkPrivacyCompliance(req *VASTRequest) error {
	// COPPA compliance
	if req.COPPA == 1 {
		// Cannot serve personalized ads to children
		req.DNT = 1
		req.UID = ""
		req.IDFA = ""
		req.GID = ""
	}

	// GDPR compliance
	if req.GDPR == 1 && req.UserConsent == "" {
		return fmt.Errorf("GDPR requires user consent")
	}

	// CCPA compliance
	if strings.HasPrefix(req.USPrivacy, "1Y") {
		// User has opted out of sale
		req.DNT = 1
	}

	return nil
}

func (h *VASTHandler) getDeviceType(model string) int {
	model = strings.ToLower(model)
	switch {
	case strings.Contains(model, "roku"):
		return 3 // Connected TV
	case strings.Contains(model, "firetv"), strings.Contains(model, "fire tv"):
		return 3
	case strings.Contains(model, "appletv"), strings.Contains(model, "apple tv"):
		return 3
	case strings.Contains(model, "chromecast"):
		return 3
	case strings.Contains(model, "smarttv"), strings.Contains(model, "smart tv"):
		return 3
	case strings.Contains(model, "iphone"), strings.Contains(model, "ipad"):
		return 4 // Phone/Tablet
	case strings.Contains(model, "android"):
		return 4
	default:
		return 2 // Connected Device
	}
}

func (h *VASTHandler) getDeviceMake(model string) string {
	model = strings.ToLower(model)
	switch {
	case strings.Contains(model, "roku"):
		return "Roku"
	case strings.Contains(model, "fire"):
		return "Amazon"
	case strings.Contains(model, "apple"):
		return "Apple"
	case strings.Contains(model, "chromecast"):
		return "Google"
	case strings.Contains(model, "samsung"):
		return "Samsung"
	case strings.Contains(model, "lg"):
		return "LG"
	default:
		return "Unknown"
	}
}

func (h *VASTHandler) getIFA(req *VASTRequest) string {
	if req.IDFA != "" {
		return req.IDFA
	}
	if req.GID != "" {
		return req.GID
	}
	if req.IFV != "" {
		return req.IFV
	}
	return req.UID
}

func (h *VASTHandler) buildTrackingURL(event string, req *VASTRequest, bid *Bid) string {
	base := "https://track.lux.network/v1/event"
	params := fmt.Sprintf("?event=%s&imp=%s&zone=%d&app=%s&bid=%s",
		event, bid.ImpID, req.ZoneID, req.AppToken, bid.ID)
	
	// Add blockchain tracking if enabled
	if req.OnChainTracking == 1 && req.WalletAddress != "" {
		params += fmt.Sprintf("&wallet=%s&chain=%d", req.WalletAddress, req.ChainID)
	}
	
	return base + params
}

func (h *VASTHandler) getMediaFilesForLayout(layout string, videoURL string) []MediaFile {
	files := []MediaFile{}
	
	switch layout {
	case "s": // Small - 320x180
		files = append(files, MediaFile{
			Delivery: "progressive",
			Type:     "video/mp4",
			Width:    320,
			Height:   180,
			Bitrate:  500,
			URL:      videoURL,
		})
	case "m": // Medium - 640x360
		files = append(files, MediaFile{
			Delivery: "progressive",
			Type:     "video/mp4",
			Width:    640,
			Height:   360,
			Bitrate:  1000,
			URL:      videoURL,
		})
	case "l": // Large - 1280x720
		files = append(files, MediaFile{
			Delivery: "progressive",
			Type:     "video/mp4",
			Width:    1280,
			Height:   720,
			Bitrate:  2500,
			URL:      videoURL,
		})
	case "xl": // Extra Large - 1920x1080
		files = append(files, MediaFile{
			Delivery: "progressive",
			Type:     "video/mp4",
			Width:    1920,
			Height:   1080,
			Bitrate:  5000,
			URL:      videoURL,
		})
	default:
		// Default to medium
		files = append(files, MediaFile{
			Delivery: "progressive",
			Type:     "video/mp4",
			Width:    640,
			Height:   360,
			Bitrate:  1000,
			URL:      videoURL,
		})
	}
	
	// Add WebM alternative for better compatibility
	for _, mp4File := range files {
		webmFile := mp4File
		webmFile.Type = "video/webm"
		webmFile.URL = strings.Replace(videoURL, ".mp4", ".webm", 1)
		files = append(files, webmFile)
	}
	
	// Add HLS for streaming
	files = append(files, MediaFile{
		Delivery: "streaming",
		Type:     "application/x-mpegURL",
		Width:    1920,
		Height:   1080,
		URL:      strings.Replace(videoURL, ".mp4", ".m3u8", 1),
	})
	
	return files
}

func (h *VASTHandler) getOMIDVerificationScript(partner string) string {
	// Return OMID verification script URL based on partner
	scripts := map[string]string{
		"iabtechlab": "https://cdn.lux.network/omid/omid-validation-verification-script-v1.js",
		"moat":       "https://cdn.lux.network/omid/moat-omid-verification.js",
		"doubleverify": "https://cdn.lux.network/omid/dv-omid-verification.js",
		"ias":        "https://cdn.lux.network/omid/ias-omid-verification.js",
	}
	
	if url, ok := scripts[strings.ToLower(partner)]; ok {
		return url
	}
	
	return "https://cdn.lux.network/omid/default-verification.js"
}

func (h *VASTHandler) trackImpression(req *VASTRequest, vast *VAST) {
	// Track impression asynchronously
	impression := &ImpressionRecord{
		ID:        uuid.New().String(),
		Timestamp: time.Now(),
		AppToken:  req.AppToken,
		ZoneID:    req.ZoneID,
		Device: DeviceInfo{
			OS:          req.OS,
			OSVersion:   req.OSVer,
			Model:       req.DeviceModel,
			IFA:         h.getIFA(req),
		},
		Location: LocationInfo{
			Lat:     req.Lat,
			Lon:     req.Long,
			Country: "", // Would be derived from IP
		},
		AdCount: len(vast.Ads),
	}
	
	// Store impression
	if err := h.Storage.StoreImpression(impression); err != nil {
		// Log error but don't fail the request
		fmt.Printf("Failed to store impression: %v\n", err)
	}
	
	// Update analytics
	h.Analytics.TrackImpression(impression)
	
	// Blockchain tracking if enabled
	if req.OnChainTracking == 1 && req.WalletAddress != "" {
		h.BlockchainMgr.RecordImpression(impression, req.WalletAddress, req.ChainID)
	}
}

func formatDuration(seconds int) string {
	hours := seconds / 3600
	minutes := (seconds % 3600) / 60
	secs := seconds % 60
	return fmt.Sprintf("%02d:%02d:%02d", hours, minutes, secs)
}

// VASTError represents an error response
type VASTError struct {
	XMLName xml.Name `xml:"Error"`
	Code    int      `xml:"code,attr"`
	Message string   `xml:",chardata"`
}
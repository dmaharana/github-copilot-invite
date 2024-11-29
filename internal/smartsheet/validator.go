package smartsheet

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
)

type LicenseValidator struct {
	client    *http.Client
	token     string
	sheetID   int64
	cacheLock sync.RWMutex
	cache     map[string]int // org -> available licenses
}

type Sheet struct {
	Rows []Row `json:"rows"`
}

type Row struct {
	Cells []Cell `json:"cells"`
}

type Cell struct {
	Value interface{} `json:"value"`
}

func NewLicenseValidator(token string, sheetID int64) *LicenseValidator {
	return &LicenseValidator{
		client:  &http.Client{},
		token:   token,
		sheetID: sheetID,
		cache:   make(map[string]int),
	}
}

func (v *LicenseValidator) RefreshLicenseCache() error {
	url := fmt.Sprintf("https://api.smartsheet.com/2.0/sheets/%d", v.sheetID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Add("Authorization", "Bearer "+v.token)
	resp, err := v.client.Do(req)
	if err != nil {
		return fmt.Errorf("error making request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var sheet Sheet
	if err := json.NewDecoder(resp.Body).Decode(&sheet); err != nil {
		return fmt.Errorf("error decoding response: %v", err)
	}

	v.cacheLock.Lock()
	defer v.cacheLock.Unlock()

	// Clear existing cache
	v.cache = make(map[string]int)

	// Process rows and update cache
	// Note: Column indices should be adjusted based on your actual sheet structure
	for _, row := range sheet.Rows {
		if len(row.Cells) < 2 {
			continue
		}
		
		org, ok := row.Cells[0].Value.(string)
		if !ok {
			continue
		}
		
		licenses, ok := row.Cells[1].Value.(float64)
		if !ok {
			continue
		}
		
		v.cache[org] = int(licenses)
	}

	return nil
}

func (v *LicenseValidator) CheckLicenseAvailability(org string) (bool, error) {
	v.cacheLock.RLock()
	licenses, exists := v.cache[org]
	v.cacheLock.RUnlock()

	if !exists {
		if err := v.RefreshLicenseCache(); err != nil {
			return false, err
		}
		v.cacheLock.RLock()
		licenses = v.cache[org]
		v.cacheLock.RUnlock()
	}

	return licenses > 0, nil
}

func (v *LicenseValidator) DecrementLicense(org string) error {
	v.cacheLock.Lock()
	defer v.cacheLock.Unlock()

	if v.cache[org] <= 0 {
		return fmt.Errorf("no licenses available for org: %s", org)
	}

	// Update local cache
	v.cache[org]--

	// Update Smartsheet
	// This would involve making a PUT request to update the specific cell
	// Implementation depends on your sheet structure and business logic
	// For now, we'll just return success
	return nil
}

package demo

import (
	"time"

	"github.com/forhsd/sqlbuilder/repository"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type MetabaseDatabase struct {
	ID                       int                `json:"id"`
	Name                     string             `json:"name"`
	CreatedAt                time.Time          `json:"created_at"`
	UpdatedAt                time.Time          `json:"updated_at"`
	Description              string             `json:"description"`
	Details                  string             `json:"details"`
	Engine                   string             `json:"engine"`
	IsSample                 repository.BitBool `json:"is_sample"`
	IsFullSync               repository.BitBool `json:"is_full_sync"`
	PointsOfInterest         string             `json:"points_of_interest"`
	Caveats                  string             `json:"caveats"`
	MetadataSyncSchedule     string             `json:"metadata_sync_schedule"`
	CacheFieldValuesSchedule string             `json:"cache_field_values_schedule"`
	Timezone                 string             `json:"timezone"`
	IsOnDemand               repository.BitBool `json:"is_on_demand"`
	Options                  string             `json:"options"`
	AutoRunQueries           repository.BitBool `json:"auto_run_queries"`
	Refingerprint            repository.BitBool `json:"refingerprint"`
	CacheTTL                 int                `json:"cache_ttl"`
	InitialSyncStatus        string             `json:"initial_sync_status"`
	CreatorID                int                `json:"creator_id"`
	Settings                 string             `json:"settings"`
	DBMSVersion              string             `json:"dbms_version"`
}

// JSONObject is a placeholder for representing JSON objects in Go
type JSONObject map[string]any

func GetOneDatabase(c *gin.Context, id int) (MetabaseDatabase, error) {
	// 从 Gin 上下文中获取 db 对象
	db := c.MustGet("db").(*gorm.DB)

	database := MetabaseDatabase{}
	err := db.Model(&MetabaseDatabase{}).Where("id = ?", id).First(&database).Error
	return database, err
}

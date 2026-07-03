package config

import (
	"PocketArtisan/internal/entities"
	"PocketArtisan/internal/modules/utils"
	"log"
	"os"
	"strings"
	"time"

	"github.com/lib/pq"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// seedItem describes a craft or product category with its human keywords.
type seedItem struct {
	Name     string
	Keywords []string
}

var craftSeed = []seedItem{
	{"Kovač", []string{"kovač", "blacksmith"}},
	{"Duborezac", []string{"duborezac", "woodcarver"}},
	{"Obućar", []string{"obućar", "shoemaker"}},
	{"Grnčar", []string{"grnčar", "potter"}},
	{"Bačvar", []string{"bačvar", "cooper"}},
	{"Stolar", []string{"stolar", "carpenter"}},
	{"Tesar", []string{"tesar", "joiner"}},
	{"Kamenorezac", []string{"kamenorezac", "stonemason"}},
	{"Juvelir", []string{"juvelir", "zlatar", "goldsmith", "jeweler"}},
	{"Časovničar", []string{"časovničar", "watchmaker"}},
	{"Krojač", []string{"krojač", "tailor"}},
	{"Korpar", []string{"korpar", "basketweaver", "basket maker"}},
	{"Staklorezac", []string{"staklorezac", "glassworker", "glassmaker"}},
	{"Keramičar", []string{"keramičar", "ceramicist", "tile setter"}},
	{"Rezbar", []string{"rezbar", "engraver"}},
	{"Ikonopisac", []string{"ikonopisac", "iconographer", "icon painter"}},
	{"Kaligraf", []string{"kaligraf", "calligrapher"}},
	{"Izrađivač nakita", []string{"nakit", "izrađivač nakita", "jewelry maker", "handmade jewelry"}},
	{"Izrađivač kožnih proizvoda", []string{"koža", "kožni proizvodi", "leatherworker", "leather craft"}},
	{"Izrađivač igračaka", []string{"igračke", "toymaker"}},
	{"Slikar", []string{"slikar", "artist", "painter"}},
	{"Vajar", []string{"vajar", "sculptor"}},
}

var productCategorySeed = []seedItem{
	{"Alati i oruđa", []string{"kovač", "blacksmith", "alat", "okovi", "čelik"}},
	{"Drveni suveniri", []string{"duborezac", "woodcarver", "drvorez", "suveniri", "ukrasi"}},
	{"Obuća", []string{"obućar", "shoemaker", "cipele", "sandale", "koža"}},
	{"Posuđe i keramika", []string{"grnčar", "potter", "keramičar", "ceramicist", "keramika", "posuđe", "vaze"}},
	{"Drvena burad i ambalaža", []string{"bačvar", "cooper", "burad", "drvena ambalaža"}},
	{"Nameštaj", []string{"stolar", "tesar", "carpenter", "joiner", "nameštaj", "drvene konstrukcije"}},
	{"Kameni elementi", []string{"kamenorezac", "stonemason", "kamen", "ploče", "spomenici"}},
	{"Nakit", []string{"juvelir", "zlatar", "jeweler", "izrađivač nakita", "prstenje", "ogrlice"}},
	{"Satovi", []string{"časovničar", "watchmaker", "ručni satovi", "zidni satovi"}},
	{"Odevni predmeti", []string{"krojač", "tailor", "odeća", "tekstil"}},
	{"Pletene korpe", []string{"korpar", "basketweaver", "korpe", "nameštaj od pruća"}},
	{"Stakleni ukrasi", []string{"staklorezac", "glassworker", "staklo", "ogledala", "vitraži"}},
	{"Umetničke gravure", []string{"rezbar", "engraver", "graviranje", "pločice"}},
	{"Ikonografija", []string{"ikonopisac", "iconographer", "ikone", "verski predmeti"}},
	{"Kaligrafija", []string{"kaligraf", "calligrapher", "ispisivanje", "sertifikati"}},
	{"Kožni proizvodi", []string{"izrađivač kožnih proizvoda", "leatherworker", "torbe", "kaiševi"}},
	{"Igračke", []string{"izrađivač igračaka", "toymaker", "drvene igračke", "ručni rad"}},
	{"Slike i skulpture", []string{"slikar", "painter", "vajar", "sculptor", "umetnost", "skulpture", "platna"}},
}

// runSeeds populates baseline reference data required for the application to be
// usable right after an initial migration: an admin user, the craftsman types
// (crafts) and the product categories. Every seed is idempotent, so it is safe
// to run on every startup.
func runSeeds() {
	log.Println("Seeding baseline data...")
	seedAdminUser()
	seedCrafts()
	seedProductCategories()
}

func buildSearchKeywords(name string, keywords []string) pq.StringArray {
	out := make(pq.StringArray, 0, len(keywords)+1)
	out = append(out, utils.NormalizeForSearch(name))
	for _, kw := range keywords {
		out = append(out, utils.NormalizeForSearch(kw))
	}
	return out
}

func seedAdminUser() {
	username := mustEnv("ADMIN_USERNAME")
	email := mustEnv("ADMIN_EMAIL")
	password := mustEnv("ADMIN_PASSWORD")

	var existing entities.User
	err := DB.Where("username = ? OR email = ?", username, email).First(&existing).Error
	if err == nil {
		return // admin already present
	}
	if err != gorm.ErrRecordNotFound {
		log.Fatalf("Failed to check for existing admin user: %v", err)
	}

	dob, _ := time.Parse("2006-01-02", "2000-01-01")
	admin := &entities.User{
		Username:       username,
		Email:          email,
		Firstname:      "Marko",
		Lastname:       "Jovanovic",
		DateOfBirth:    dob,
		Gender:         "male",
		Role:           "admin",
		ProfilePicture: adminAvatarURL(),
	}
	if err := admin.SetPassword(password); err != nil {
		log.Fatalf("Failed to hash admin password: %v", err)
	}

	err = DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(admin).Error; err != nil {
			return err
		}
		return tx.Create(&entities.Cart{UserID: admin.ID}).Error
	})
	if err != nil {
		log.Fatalf("Failed to seed admin user: %v", err)
	}
	log.Printf("Seeded admin user %q", username)
}

func seedCrafts() {
	for _, item := range craftSeed {
		craft := entities.Craft{
			Name:           item.Name,
			Keywords:       pq.StringArray(item.Keywords),
			SearchKeywords: buildSearchKeywords(item.Name, item.Keywords),
		}
		if err := DB.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "name"}},
			DoNothing: true,
		}).Create(&craft).Error; err != nil {
			log.Fatalf("Failed to seed craft %q: %v", item.Name, err)
		}
	}
	log.Printf("Seeded %d crafts", len(craftSeed))
}

func seedProductCategories() {
	for _, item := range productCategorySeed {
		category := entities.ProductCategory{
			Name:           item.Name,
			Keywords:       pq.StringArray(item.Keywords),
			SearchKeywords: buildSearchKeywords(item.Name, item.Keywords),
		}
		if err := DB.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "name"}},
			DoNothing: true,
		}).Create(&category).Error; err != nil {
			log.Fatalf("Failed to seed product category %q: %v", item.Name, err)
		}
	}
	log.Printf("Seeded %d product categories", len(productCategorySeed))
}

// adminAvatarURL mirrors the default avatar resolution used at registration:
// it points to R2 when configured, otherwise the local asset route.
func adminAvatarURL() string {
	base := "http://localhost:8080/api/assets/avatars"
	if publicURL := os.Getenv("R2_PUBLIC_URL"); publicURL != "" {
		base = strings.TrimRight(publicURL, "/") + "/avatars"
	}
	return base + "/default_male.png"
}

func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		log.Fatalf("%s must be set to seed the admin user", key)
	}
	return v
}

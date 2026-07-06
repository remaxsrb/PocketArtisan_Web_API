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

type seedItem struct {
	Name     string
	Keywords []string
}

var craftSeed = []seedItem{
	{"Ковач", []string{"kovač", "ковач", "blacksmith"}},
	{"Дуборезац", []string{"duborezac", "дуборезац", "woodcarver"}},
	{"Обућар", []string{"obućar", "обућар", "shoemaker"}},
	{"Грнчар", []string{"grnčar", "грнчар", "potter"}},
	{"Бачвар", []string{"bačvar", "бачвар", "cooper"}},
	{"Столар", []string{"stolar", "столар", "carpenter"}},
	{"Тесар", []string{"tesar", "тесар", "joiner"}},
	{"Каменорезац", []string{"kamenorezac", "каменорезац", "stonemason"}},
	{"Јувелир", []string{"juvelir", "јувелир", "zlatar", "златар", "goldsmith", "jeweler"}},
	{"Часовничар", []string{"časovničar", "часовничар", "watchmaker"}},
	{"Кројач", []string{"krojač", "кројач", "tailor"}},
	{"Корпар", []string{"korpar", "корпар", "basketweaver", "basket maker"}},
	{"Стаклорезац", []string{"staklorezac", "стаклорезац", "glassworker", "glassmaker"}},
	{"Керамичар", []string{"keramičar", "керамичар", "ceramicist", "tile setter"}},
	{"Резбар", []string{"rezbar", "резбар", "engraver"}},
	{"Иконописац", []string{"ikonopisac", "иконописац", "iconographer", "icon painter"}},
	{"Калиграф", []string{"kaligraf", "калиграф", "calligrapher"}},
	{"Израђивач кожних производа", []string{"koža", "кожа", "kožni proizvodi", "кожни производи", "leatherworker", "leather craft"}},
	{"Израђивач играчака", []string{"igračke", "играчке", "toymaker"}},
	{"Сликар", []string{"slikar", "сликар", "artist", "painter"}},
	{"Вајар", []string{"vajar", "вајар", "sculptor"}},
}

var productCategorySeed = []seedItem{
	{"Алати и оружја", []string{"kovač", "ковач", "blacksmith", "alat", "алат", "okovi", "окови", "čelik", "челик"}},
	{"Дрвени сувенири", []string{"duborezac", "дуборезац", "woodcarver", "drvorez", "дрворез", "suveniri", "сувенири", "ukrasi", "украси"}},
	{"Обућа", []string{"obućar", "обућар", "shoemaker", "cipele", "ципеле", "sandale", "сандале", "koža", "кожа"}},
	{"Посуђе и керамика", []string{"grnčar", "грнчар", "potter", "keramičar", "керамичар", "ceramicist", "keramika", "керамика", "posuđe", "посуђе", "vaze", "вазе"}},
	{"Дрвена бурад и амбалажа", []string{"bačvar", "бачвар", "cooper", "burad", "бурад", "drvena ambalaža", "дрвена амбалажа"}},
	{"Намештај", []string{"stolar", "столар", "tesar", "тесар", "carpenter", "joiner", "nameštaj", "намештај", "drvene konstrukcije", "дрвене конструкције"}},
	{"Камени елементи", []string{"kamenorezac", "каменорезац", "stonemason", "kamen", "камен", "ploče", "плоче", "spomenici", "споменици"}},
	{"Огрлице", []string{"juvelir", "јувелир", "zlatar", "огрлице", "jeweler", "izrađivač nakita", "израђивач накита", "prstenje", "прстење", "ogrlice", "огрлице"}},
	{"Прстење", []string{"juvelir", "јувелир", "zlatar", "прстење", "jeweler", "izrađivač nakita", "израђивач накита", "prstenje", "прстење", "ogrlice", "огрлице"}},
	{"Минђуше", []string{"juvelir", "јувелир", "zlatar", "минђуше", "jeweler", "izrađivač nakita", "израђивач накита", "prstenje", "прстење", "ogrlice", "огрлице"}},

	{"Сатови", []string{"časovničar", "часовничар", "watchmaker", "ručni satovi", "ручни сатови", "zidni satovi", "зидни сатови"}},
	{"Одевни предмети", []string{"krojač", "кројач", "tailor", "odeća", "одећа", "tekstil", "текстил"}},
	{"Плетене корпе", []string{"korpar", "корпар", "basketweaver", "korpe", "корпе", "nameštaj od pruća", "намештај од прућа"}},
	{"Стаклени украси", []string{"staklorezac", "стаклорезац", "glassworker", "staklo", "стакло", "ogledala", "огледала", "vitraži", "витражи"}},
	{"Уметничке гравуре", []string{"rezbar", "резбар", "engraver", "graviranje", "гравирање", "pločice", "плочице"}},
	{"Иконографија", []string{"ikonopisac", "иконописац", "iconographer", "ikone", "иконе", "verski predmeti", "верски предмети"}},
	{"Калиграфија", []string{"kaligraf", "калиграф", "calligrapher", "ispisivanje", "исписивање", "sertifikati", "сертификати"}},
	{"Кожни производи", []string{"izrađivač kožnih proizvoda", "израђивач кожних производа", "leatherworker", "torbe", "торбе", "kaiševi", "каишеви"}},
	{"Играчке", []string{"izrađivač igračaka", "израђивач играчака", "toymaker", "drvene igračke", "дрвене играчке", "ručni rad", "ручни рад"}},
	{"Слике и скулптуре", []string{"slikar", "сликар", "painter", "vajar", "вајар", "sculptor", "umetnost", "уметност", "skulpture", "скулптуре", "platna", "платна"}},
}

var craftCategoryLinks = map[string][]string{
	"Ковач":       {"Алати и оружја"},
	"Дуборезац":   {"Дрвени сувенири"},
	"Обућар":      {"Обућа"},
	"Грнчар":      {"Посуђе и керамика"},
	"Бачвар":      {"Дрвена бурад и амбалажа"},
	"Столар":      {"Намештај"},
	"Тесар":       {"Намештај"},
	"Каменорезац": {"Камени елементи"},
	"Јувелир":     {"Прстење", "Минђуше", "Огрлице"},
	"Часовничар":  {"Сатови"},
	"Кројач":      {"Одевни предмети"},
	"Корпар":      {"Плетене корпе"},
	"Стаклорезац": {"Стаклени украси"},
	"Керамичар":   {"Посуђе и керамика"},
	"Резбар":      {"Уметничке гравуре"},
	"Иконописац":  {"Иконографија"},
	"Калиграф":    {"Калиграфија"},
	"Израђивач кожних производа": {"Кожни производи"},
	"Израђивач играчака":         {"Играчке"},
	"Сликар":                     {"Слике и скулптуре"},
	"Вајар":                      {"Слике и скулптуре"},
}

func runSeeds() {
	log.Println("Seeding baseline data...")
	seedAdminUser()
	seedCrafts()
	seedProductCategories()
	seedCraftProductCategories()
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

func seedCraftProductCategories() {
	var links []entities.CraftProductCategory
	for craftName, categoryNames := range craftCategoryLinks {
		var craft entities.Craft
		if err := DB.Where("name = ?", craftName).First(&craft).Error; err != nil {
			log.Fatalf("Failed to look up craft %q for category linking: %v", craftName, err)
		}
		for _, categoryName := range categoryNames {
			var category entities.ProductCategory
			if err := DB.Where("name = ?", categoryName).First(&category).Error; err != nil {
				log.Fatalf("Failed to look up product category %q for craft %q: %v", categoryName, craftName, err)
			}
			links = append(links, entities.CraftProductCategory{CraftID: craft.ID, CategoryID: category.ID})
		}
	}

	if err := DB.Clauses(clause.OnConflict{DoNothing: true}).Create(&links).Error; err != nil {
		log.Fatalf("Failed to seed craft-product category links: %v", err)
	}
	log.Printf("Seeded %d craft-product category links", len(links))
}

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

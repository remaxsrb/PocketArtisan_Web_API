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
	{"Ножеви", []string{"noževi", "ножеви", "knives", "sečiva", "сечива"}},
	{"Мачеви", []string{"mačevi", "мачеви", "swords"}},
	{"Ковани украси и ограде", []string{"kovani ukrasi", "ковани украси", "ograde", "ограде", "kapije", "капије", "wrought iron"}},
	{"Алати", []string{"alati", "алати", "tools"}},

	{"Дрвене фигурице", []string{"drvene figurice", "дрвене фигурице", "wooden figurines"}},
	{"Дрворезбарени украси", []string{"drvorezbareni ukrasi", "дрворезбарени украси", "carved ornaments"}},

	{"Ципеле", []string{"cipele", "ципеле", "shoes"}},
	{"Чизме", []string{"čizme", "чизме", "boots"}},
	{"Сандале", []string{"sandale", "сандале", "sandals"}},
	{"Папуче", []string{"papuče", "папуче", "slippers"}},
	{"Патике", []string{"patike", "патике", "sneakers"}},

	{"Посуђе", []string{"posuđe", "посуђе", "dishware", "tanjiri", "тањири"}},
	{"Вазе", []string{"vaze", "вазе", "vases"}},
	{"Плочице", []string{"pločice", "плочице", "tiles"}},

	{"Буради", []string{"burad", "бурад", "barrels"}},
	{"Каце", []string{"kace", "каце", "wooden tubs", "vats"}},

	{"Столови", []string{"stolovi", "столови", "tables"}},
	{"Столице", []string{"stolice", "столице", "chairs"}},
	{"Ормани", []string{"ormani", "ормани", "wardrobes", "cabinets"}},
	{"Кревети", []string{"kreveti", "кревети", "beds"}},
	{"Полице", []string{"police", "полице", "shelves"}},
	{"Врата и прозори", []string{"vrata", "врата", "prozori", "прозори", "doors", "windows"}},

	{"Надгробни споменици", []string{"nadgrobni spomenici", "надгробни споменици", "headstones", "monuments"}},
	{"Камене плоче", []string{"kamene ploče", "камене плоче", "stone slabs", "countertops"}},
	{"Камени украси", []string{"kameni ukrasi", "камени украси", "decorative stone"}},

	{"Прстење", []string{"juvelir", "јувелир", "zlatar", "prstenje", "прстење", "jeweler", "izrađivač nakita", "израђивач накита", "rings"}},
	{"Минђуше", []string{"juvelir", "јувелир", "zlatar", "minđuše", "минђуше", "jeweler", "izrađivač nakita", "израђивач накита", "earrings"}},
	{"Огрлице", []string{"juvelir", "јувелир", "zlatar", "ogrlice", "огрлице", "jeweler", "izrađivač nakita", "израђивач накита", "necklaces"}},
	{"Наруквице", []string{"juvelir", "јувелир", "zlatar", "narukvice", "наруквице", "jeweler", "bracelets"}},

	{"Ручни сатови", []string{"ručni satovi", "ручни сатови", "wristwatches"}},
	{"Зидни сатови", []string{"zidni satovi", "зидни сатови", "wall clocks"}},
	{"Џепни сатови", []string{"džepni satovi", "џепни сатови", "pocket watches"}},

	{"Кошуље", []string{"košulje", "кошуље", "shirts"}},
	{"Панталоне", []string{"pantalone", "панталоне", "trousers"}},
	{"Хаљине", []string{"haljine", "хаљине", "dresses"}},
	{"Костими", []string{"kostimi", "костими", "suits"}},

	{"Корпе за пикник", []string{"korpe za piknik", "корпе за пикник", "picnic baskets"}},
	{"Плетени намештај", []string{"pleteni nameštaj", "плетени намештај", "wicker furniture"}},
	{"Декоративне корпе", []string{"dekorativne korpe", "декоративне корпе", "decorative baskets"}},

	{"Витражи", []string{"vitraži", "витражи", "stained glass"}},
	{"Огледала", []string{"ogledala", "огледала", "mirrors"}},
	{"Стаклене вазе", []string{"staklene vaze", "стаклене вазе", "glass vases"}},

	{"Гравиране плочице", []string{"gravirane pločice", "гравиране плочице", "engraved plaques"}},
	{"Персонализовани поклони", []string{"personalizovani pokloni", "персонализовани поклони", "personalized gifts"}},

	{"Иконе", []string{"ikone", "иконе", "icons"}},
	{"Верски украси", []string{"verski ukrasi", "верски украси", "religious ornaments"}},

	{"Сертификати и дипломе", []string{"sertifikati", "сертификати", "diplome", "дипломе", "certificates"}},
	{"Позивнице", []string{"pozivnice", "позивнице", "invitations"}},

	{"Торбе", []string{"torbe", "торбе", "bags"}},
	{"Каишеви", []string{"kaiševi", "каишеви", "belts"}},
	{"Новчаници", []string{"novčanici", "новчаници", "wallets"}},

	{"Дрвене играчке", []string{"drvene igračke", "дрвене играчке", "wooden toys"}},
	{"Лутке", []string{"lutke", "лутке", "dolls"}},
	{"Друштвене игре", []string{"društvene igre", "друштвене игре", "board games"}},

	{"Слике", []string{"slike", "слике", "paintings"}},
	{"Скулптуре", []string{"skulpture", "скулптуре", "sculptures"}},
}

var craftCategoryLinks = map[string][]string{
	"Ковач":       {"Ножеви", "Мачеви", "Ковани украси и ограде", "Алати"},
	"Дуборезац":   {"Дрвене фигурице", "Дрворезбарени украси"},
	"Обућар":      {"Ципеле", "Чизме", "Сандале", "Папуче", "Патике"},
	"Грнчар":      {"Посуђе", "Вазе"},
	"Керамичар":   {"Посуђе", "Плочице"},
	"Бачвар":      {"Буради", "Каце"},
	"Столар":      {"Столови", "Столице", "Ормани", "Кревети", "Полице"},
	"Тесар":       {"Столови", "Ормани", "Врата и прозори"},
	"Каменорезац": {"Надгробни споменици", "Камене плоче", "Камени украси"},
	"Јувелир":     {"Прстење", "Минђуше", "Огрлице", "Наруквице"},
	"Часовничар":  {"Ручни сатови", "Зидни сатови", "Џепни сатови"},
	"Кројач":      {"Кошуље", "Панталоне", "Хаљине", "Костими"},
	"Корпар":      {"Корпе за пикник", "Плетени намештај", "Декоративне корпе"},
	"Стаклорезац": {"Витражи", "Огледала", "Стаклене вазе"},
	"Резбар":      {"Гравиране плочице", "Персонализовани поклони"},
	"Иконописац":  {"Иконе", "Верски украси"},
	"Калиграф":    {"Сертификати и дипломе", "Позивнице"},
	"Израђивач кожних производа": {"Торбе", "Каишеви", "Новчаници"},
	"Израђивач играчака":         {"Дрвене играчке", "Лутке", "Друштвене игре"},
	"Сликар":                     {"Слике"},
	"Вајар":                      {"Скулптуре"},
}

func runSeeds() {
	// Schema and reference data (crafts, product categories, links) are now
	// applied by versioned migrations (see ./migrations and cmd/seedgen). Only
	// the admin user is seeded here because it depends on runtime env vars and
	// a bcrypt-hashed password, which don't belong in a static SQL migration.
	log.Println("Seeding baseline data...")
	seedAdminUser()
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
		// Upsert on name so keyword changes in code are refreshed on every
		// deploy while the row's primary key (and any craftsmen FK) is preserved.
		if err := DB.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "name"}},
			DoUpdates: clause.AssignmentColumns([]string{"keywords", "search_keywords"}),
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
		// Upsert on name so keyword changes in code are refreshed on every
		// deploy while the row's primary key (and any product FK) is preserved.
		if err := DB.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "name"}},
			DoUpdates: clause.AssignmentColumns([]string{"keywords", "search_keywords"}),
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

	// Fully rebuild the join table on every deploy so the mapping in code is the
	// single source of truth: this erases stale/removed links (which otherwise
	// leave a craft with no categories) and re-inserts the current set atomically.
	err := DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Session(&gorm.Session{AllowGlobalUpdate: true}).
			Delete(&entities.CraftProductCategory{}).Error; err != nil {
			return err
		}
		if len(links) == 0 {
			return nil
		}
		return tx.Create(&links).Error
	})
	if err != nil {
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

package repositories

import (
	"errors"
	"metachan/entities"
	"metachan/enums"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func GetAllCharacterStubs() ([]characterStub, error) {
	var stubs []characterStub
	if err := DB.Model(&entities.Character{}).Select("mal_id, enriched_at").Scan(&stubs).Error; err != nil {
		return nil, err
	}
	return stubs, nil
}

func UpdateCharacterDetails(malID int, name, nameKanji, url, imageURL, about string, nicknames []string, favorites int, voiceActors []entities.CharacterVoiceActor, animeAppearances []entities.CharacterAnimeAppearance) error {
	var char entities.Character
	if err := DB.Where("mal_id = ?", malID).First(&char).Error; err != nil {
		return err
	}

	char.Name = name
	char.NameKanji = nameKanji
	char.URL = url
	char.ImageURL = imageURL
	char.About = about
	char.Nicknames = nicknames
	char.Favorites = favorites

	if err := DB.Save(&char).Error; err != nil {
		return err
	}

	return DB.Transaction(func(tx *gorm.DB) error {
		tx.Where("character_id = ?", char.ID).Delete(&entities.CharacterVoiceActor{})
		for _, cva := range voiceActors {
			va := cva.Person
			if va == nil {
				continue
			}

			tx.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "mal_id"}},
				DoUpdates: clause.AssignmentColumns([]string{"url", "image", "name"}),
			}).Create(va)
			if va.ID == 0 {
				tx.Where("mal_id = ?", va.MALID).First(va)
			}

			tx.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "character_id"}, {Name: "person_id"}},
				DoUpdates: clause.AssignmentColumns([]string{"language"}),
			}).Create(&entities.CharacterVoiceActor{
				CharacterID: char.ID,
				PersonID:    va.ID,
				Language:    cva.Language,
			})
		}

		tx.Where("character_id = ?", char.ID).Delete(&entities.CharacterAnimeAppearance{})
		for i := range animeAppearances {
			animeAppearances[i].CharacterID = char.ID
		}
		if len(animeAppearances) > 0 {
			tx.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "character_id"}, {Name: "anime_mal_id"}},
				DoUpdates: clause.AssignmentColumns([]string{"title", "url", "image_url", "role"}),
			}).Create(&animeAppearances)
		}

		return nil
	})
}

func GetAnimeCharacters[T idType](maptype enums.MappingType, id T) ([]entities.Character, error) {
	mapping, err := GetAnimeMapping(maptype, id)
	if err != nil {
		return nil, errors.New("anime not found")
	}

	var anime entities.Anime
	if err := DB.Where("mapping_id = ?", mapping.ID).Select("id").First(&anime).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		return nil, errors.New("anime not found")
	}

	var rows []struct {
		CharacterID uint
		Role        string
	}
	DB.Table("anime_characters").
		Select("character_id, role").
		Where("anime_id = ?", anime.ID).
		Scan(&rows)

	if len(rows) == 0 {
		return []entities.Character{}, nil
	}

	charIDs := make([]uint, len(rows))
	roleMap := make(map[uint]string, len(rows))
	for i, row := range rows {
		charIDs[i] = row.CharacterID
		roleMap[row.CharacterID] = row.Role
	}

	var characters []entities.Character
	DB.Where("id IN ?", charIDs).Find(&characters)

	var voiceActors []entities.CharacterVoiceActor
	DB.Preload("Person").Where("character_id IN ?", charIDs).Find(&voiceActors)

	voiceActorsByCharacterID := make(map[uint][]entities.CharacterVoiceActor)
	for _, voiceActor := range voiceActors {
		voiceActorsByCharacterID[voiceActor.CharacterID] = append(voiceActorsByCharacterID[voiceActor.CharacterID], voiceActor)
	}

	for i := range characters {
		characters[i].Role = roleMap[characters[i].ID]
		characters[i].VoiceActors = voiceActorsByCharacterID[characters[i].ID]
	}

	return characters, nil
}

func GetAnimeCharacter[T idType](maptype enums.MappingType, id T, characterMALID int) (entities.Character, error) {
	mapping, err := GetAnimeMapping(maptype, id)
	if err != nil {
		return entities.Character{}, errors.New("anime not found")
	}

	var anime entities.Anime
	if err := DB.Where("mapping_id = ?", mapping.ID).Select("id").First(&anime).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return entities.Character{}, err
		}
		return entities.Character{}, errors.New("anime not found")
	}

	var char entities.Character
	if err := DB.
		Preload("VoiceActors.Person").
		Preload("AnimeAppearances").
		Where("mal_id = ?", characterMALID).
		First(&char).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return entities.Character{}, err
		}
		return entities.Character{}, errors.New("character not found")
	}

	var ac animeCharacterRow
	if err := DB.Table("anime_characters").
		Select("role").
		Where("anime_id = ? AND character_id = ?", anime.ID, char.ID).
		Scan(&ac).Error; err == nil {
		char.Role = ac.Role
	}

	return char, nil
}

func loadAnimeCharacters(anime *entities.Anime) {
	var rows []struct {
		CharacterID uint
		Role        string
	}
	DB.Table("anime_characters").
		Select("character_id, role").
		Where("anime_id = ?", anime.ID).
		Scan(&rows)

	if len(rows) == 0 {
		return
	}

	charIDs := make([]uint, len(rows))
	roleMap := make(map[uint]string, len(rows))
	for i, row := range rows {
		charIDs[i] = row.CharacterID
		roleMap[row.CharacterID] = row.Role
	}

	var characters []entities.Character
	DB.Where("id IN ?", charIDs).Find(&characters)

	var voiceActors []entities.CharacterVoiceActor
	DB.Preload("Person").Where("character_id IN ?", charIDs).Find(&voiceActors)

	voiceActorsByCharacterID := make(map[uint][]entities.CharacterVoiceActor)
	for _, voiceActor := range voiceActors {
		voiceActorsByCharacterID[voiceActor.CharacterID] = append(voiceActorsByCharacterID[voiceActor.CharacterID], voiceActor)
	}

	for i := range characters {
		characters[i].Role = roleMap[characters[i].ID]
		characters[i].VoiceActors = voiceActorsByCharacterID[characters[i].ID]
	}

	anime.Characters = characters
}

func SaveAnimeCharacters(animeID uint, characters []entities.Character) error {
	for i := range characters {
		char := &characters[i]

		DB.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "mal_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"url", "image_url", "name"}),
		}).Create(char)
		if char.ID == 0 {
			DB.Where("mal_id = ?", char.MALID).First(char)
		}

		DB.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "anime_id"}, {Name: "character_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"role"}),
		}).Create(&entities.AnimeCharacter{
			AnimeID:     animeID,
			CharacterID: char.ID,
			Role:        char.Role,
		})

		for _, cva := range char.VoiceActors {
			va := cva.Person
			if va == nil {
				continue
			}

			DB.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "mal_id"}},
				DoUpdates: clause.AssignmentColumns([]string{"url", "image", "name"}),
			}).Create(va)
			if va.ID == 0 {
				DB.Where("mal_id = ?", va.MALID).First(va)
			}

			DB.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "character_id"}, {Name: "person_id"}},
				DoUpdates: clause.AssignmentColumns([]string{"language"}),
			}).Create(&entities.CharacterVoiceActor{
				CharacterID: char.ID,
				PersonID:    va.ID,
				Language:    cva.Language,
			})
		}
	}
	return nil
}

func SetCharacterEnriched(malID int) error {
	now := time.Now()
	return DB.Model(&entities.Character{}).Where("mal_id = ?", malID).Update("enriched_at", now).Error
}

func GetCharacterByMALID(malID int) (entities.Character, error) {
	var char entities.Character
	if err := DB.
		Preload("VoiceActors.Person").
		Preload("AnimeAppearances").
		Where("mal_id = ?", malID).
		First(&char).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return entities.Character{}, errors.New("character not found")
		}
		return entities.Character{}, err
	}
	return char, nil
}

func GetAllPersonStubs() ([]personStub, error) {
	var stubs []personStub
	if err := DB.Model(&entities.Person{}).Select("mal_id, enriched_at").Scan(&stubs).Error; err != nil {
		return nil, err
	}
	return stubs, nil
}

func UpdatePersonDetails(
	malID int,
	url, websiteURL, image, name, givenName, familyName string,
	alternateNames []string,
	birthday *time.Time,
	favorites int,
	about string,
	voiceRoles []entities.PersonVoiceRole,
	animeCredits []entities.PersonAnimeCredit,
	mangaCredits []entities.PersonMangaCredit,
) error {
	var p entities.Person
	if err := DB.Where("mal_id = ?", malID).First(&p).Error; err != nil {
		return err
	}

	p.URL = url
	p.WebsiteURL = websiteURL
	p.Image = image
	p.Name = name
	p.GivenName = givenName
	p.FamilyName = familyName
	p.AlternateNames = alternateNames
	p.Birthday = birthday
	p.Favorites = favorites
	p.About = about

	if err := DB.Save(&p).Error; err != nil {
		return err
	}

	return DB.Transaction(func(tx *gorm.DB) error {
		tx.Where("person_id = ?", p.ID).Delete(&entities.PersonVoiceRole{})
		for i := range voiceRoles {
			voiceRoles[i].PersonID = p.ID
		}
		if len(voiceRoles) > 0 {
			if err := tx.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "person_id"}, {Name: "anime_mal_id"}, {Name: "character_mal_id"}},
				DoUpdates: clause.AssignmentColumns([]string{"role", "anime_title", "anime_url", "anime_image_url", "character_name", "character_url", "character_image_url"}),
			}).Create(&voiceRoles).Error; err != nil {
				return err
			}
		}

		tx.Where("person_id = ?", p.ID).Delete(&entities.PersonAnimeCredit{})
		for i := range animeCredits {
			animeCredits[i].PersonID = p.ID
		}
		if len(animeCredits) > 0 {
			if err := tx.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "person_id"}, {Name: "anime_mal_id"}},
				DoUpdates: clause.AssignmentColumns([]string{"position", "anime_title", "anime_url", "anime_image_url"}),
			}).Create(&animeCredits).Error; err != nil {
				return err
			}
		}

		tx.Where("person_id = ?", p.ID).Delete(&entities.PersonMangaCredit{})
		for i := range mangaCredits {
			mangaCredits[i].PersonID = p.ID
		}
		if len(mangaCredits) > 0 {
			if err := tx.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "person_id"}, {Name: "manga_mal_id"}},
				DoUpdates: clause.AssignmentColumns([]string{"position", "manga_title", "manga_url", "manga_image_url"}),
			}).Create(&mangaCredits).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

func SetPersonEnriched(malID int) error {
	now := time.Now()
	return DB.Model(&entities.Person{}).Where("mal_id = ?", malID).Update("enriched_at", now).Error
}

func GetAnimePeople[T idType](maptype enums.MappingType, id T) ([]entities.Person, error) {
	mapping, err := GetAnimeMapping(maptype, id)
	if err != nil {
		return nil, errors.New("anime not found")
	}

	var anime entities.Anime
	if err := DB.Where("mapping_id = ?", mapping.ID).Select("id").First(&anime).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("anime not found")
		}
		return nil, errors.New("anime not found")
	}

	var charIDs []uint
	DB.Table("anime_characters").
		Select("character_id").
		Where("anime_id = ?", anime.ID).
		Pluck("character_id", &charIDs)

	if len(charIDs) == 0 {
		return []entities.Person{}, nil
	}

	var characters []entities.Character
	DB.Where("id IN ?", charIDs).Find(&characters)

	characterByID := make(map[uint]*entities.Character, len(characters))
	for i := range characters {
		characterByID[characters[i].ID] = &characters[i]
	}

	var characterVoiceActors []entities.CharacterVoiceActor
	DB.Preload("Person").Where("character_id IN ?", charIDs).Find(&characterVoiceActors)

	personMap := make(map[uint]*entities.Person)
	personCharacters := make(map[uint][]entities.PersonCharacterEntry)

	for _, voiceActorEntry := range characterVoiceActors {
		if voiceActorEntry.Person == nil {
			continue
		}
		if _, exists := personMap[voiceActorEntry.PersonID]; !exists {
			personMap[voiceActorEntry.PersonID] = voiceActorEntry.Person
		}
		if character, ok := characterByID[voiceActorEntry.CharacterID]; ok {
			personCharacters[voiceActorEntry.PersonID] = append(personCharacters[voiceActorEntry.PersonID], entities.PersonCharacterEntry{
				Character: character,
				Language:  voiceActorEntry.Language,
			})
		}
	}

	result := make([]entities.Person, 0, len(personMap))
	for personID, person := range personMap {
		person.Characters = personCharacters[personID]
		result = append(result, *person)
	}
	return result, nil
}

func GetPerson(malID int) (entities.Person, error) {
	var p entities.Person
	if err := DB.
		Preload("VoiceRoles").
		Preload("AnimeCredits").
		Preload("MangaCredits").
		Where("mal_id = ?", malID).
		First(&p).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return entities.Person{}, errors.New("person not found")
		}
		return entities.Person{}, err
	}
	return p, nil
}

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

	DB.Where("character_id = ?", char.ID).Delete(&entities.CharacterVoiceActor{})
	for _, cva := range voiceActors {
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

	DB.Where("character_id = ?", char.ID).Delete(&entities.CharacterAnimeAppearance{})
	for i := range animeAppearances {
		animeAppearances[i].CharacterID = char.ID
	}
	if len(animeAppearances) > 0 {
		DB.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "character_id"}, {Name: "anime_mal_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"title", "url", "image_url", "role"}),
		}).Create(&animeAppearances)
	}

	return nil
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

	var characters []entities.Character
	for _, row := range rows {
		var char entities.Character
		if err := DB.First(&char, row.CharacterID).Error; err != nil {
			continue
		}
		DB.Preload("Person").Where("character_id = ?", char.ID).Find(&char.VoiceActors)
		char.Role = row.Role
		characters = append(characters, char)
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

	for _, row := range rows {
		var char entities.Character
		if err := DB.First(&char, row.CharacterID).Error; err != nil {
			continue
		}
		DB.Preload("Person").
			Where("character_id = ?", char.ID).
			Find(&char.VoiceActors)
		char.Role = row.Role
		anime.Characters = append(anime.Characters, char)
	}
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

	DB.Where("person_id = ?", p.ID).Delete(&entities.PersonVoiceRole{})
	for i := range voiceRoles {
		voiceRoles[i].PersonID = p.ID
	}
	if len(voiceRoles) > 0 {
		DB.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "person_id"}, {Name: "anime_mal_id"}, {Name: "character_mal_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"role", "anime_title", "anime_url", "anime_image_url", "character_name", "character_url", "character_image_url"}),
		}).Create(&voiceRoles)
	}

	DB.Where("person_id = ?", p.ID).Delete(&entities.PersonAnimeCredit{})
	for i := range animeCredits {
		animeCredits[i].PersonID = p.ID
	}
	if len(animeCredits) > 0 {
		DB.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "person_id"}, {Name: "anime_mal_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"position", "anime_title", "anime_url", "anime_image_url"}),
		}).Create(&animeCredits)
	}

	DB.Where("person_id = ?", p.ID).Delete(&entities.PersonMangaCredit{})
	for i := range mangaCredits {
		mangaCredits[i].PersonID = p.ID
	}
	if len(mangaCredits) > 0 {
		DB.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "person_id"}, {Name: "manga_mal_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"position", "manga_title", "manga_url", "manga_image_url"}),
		}).Create(&mangaCredits)
	}

	return nil
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

	var charRows []struct {
		CharacterID uint
	}
	DB.Table("anime_characters").
		Select("character_id").
		Where("anime_id = ?", anime.ID).
		Scan(&charRows)

	personMap := make(map[uint]*entities.Person)
	personChars := make(map[uint][]entities.PersonCharacterEntry)

	for _, row := range charRows {
		var char entities.Character
		if err := DB.First(&char, row.CharacterID).Error; err != nil {
			continue
		}

		var cvas []entities.CharacterVoiceActor
		DB.Preload("Person").Where("character_id = ?", char.ID).Find(&cvas)

		for _, cva := range cvas {
			if cva.Person == nil {
				continue
			}
			pID := cva.PersonID
			if _, exists := personMap[pID]; !exists {
				personMap[pID] = cva.Person
			}
			charCopy := char
			personChars[pID] = append(personChars[pID], entities.PersonCharacterEntry{
				Character: &charCopy,
				Language:  cva.Language,
			})
		}
	}

	result := make([]entities.Person, 0, len(personMap))
	for pID, p := range personMap {
		p.Characters = personChars[pID]
		result = append(result, *p)
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

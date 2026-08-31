package util

// LocaleData represents locale information including language, country, and display names.
type LocaleData struct {
	LocaleString    string
	Country         string
	DisplayCountry  string
	DisplayLanguage string
	DisplayName     string
	DisplayVariant  string
	ISO3Country     string
	ISO3Language    string
	Language        string
	Variant         string
}

const (
	DefaultLanguage = "en"
)

// TODO: read from api for a dynamic list
var Locales = []LocaleData{
	{LocaleString: "ar", Country: "", DisplayCountry: "", DisplayLanguage: "Arabic", DisplayName: "Arabic", DisplayVariant: "", ISO3Country: "", ISO3Language: "ara", Language: "ar", Variant: ""},
	{LocaleString: "ar-AE", Country: "AE", DisplayCountry: "United Arab Emirates", DisplayLanguage: "Arabic", DisplayName: "Arabic (United Arab Emirates)", DisplayVariant: "", ISO3Country: "ARE", ISO3Language: "ara", Language: "ar", Variant: ""},
	{LocaleString: "ar-BH", Country: "BH", DisplayCountry: "Bahrain", DisplayLanguage: "Arabic", DisplayName: "Arabic (Bahrain)", DisplayVariant: "", ISO3Country: "BHR", ISO3Language: "ara", Language: "ar", Variant: ""},
	{LocaleString: "ar-DZ", Country: "DZ", DisplayCountry: "Algeria", DisplayLanguage: "Arabic", DisplayName: "Arabic (Algeria)", DisplayVariant: "", ISO3Country: "DZA", ISO3Language: "ara", Language: "ar", Variant: ""},
	{LocaleString: "ar-EG", Country: "EG", DisplayCountry: "Egypt", DisplayLanguage: "Arabic", DisplayName: "Arabic (Egypt)", DisplayVariant: "", ISO3Country: "EGY", ISO3Language: "ara", Language: "ar", Variant: ""},
	{LocaleString: "ar-IQ", Country: "IQ", DisplayCountry: "Iraq", DisplayLanguage: "Arabic", DisplayName: "Arabic (Iraq)", DisplayVariant: "", ISO3Country: "IRQ", ISO3Language: "ara", Language: "ar", Variant: ""},
	{LocaleString: "ar-JO", Country: "JO", DisplayCountry: "Jordan", DisplayLanguage: "Arabic", DisplayName: "Arabic (Jordan)", DisplayVariant: "", ISO3Country: "JOR", ISO3Language: "ara", Language: "ar", Variant: ""},
	{LocaleString: "ar-KW", Country: "KW", DisplayCountry: "Kuwait", DisplayLanguage: "Arabic", DisplayName: "Arabic (Kuwait)", DisplayVariant: "", ISO3Country: "KWT", ISO3Language: "ara", Language: "ar", Variant: ""},
	{LocaleString: "ar-LB", Country: "LB", DisplayCountry: "Lebanon", DisplayLanguage: "Arabic", DisplayName: "Arabic (Lebanon)", DisplayVariant: "", ISO3Country: "LBN", ISO3Language: "ara", Language: "ar", Variant: ""},
	{LocaleString: "ar-LY", Country: "LY", DisplayCountry: "Libya", DisplayLanguage: "Arabic", DisplayName: "Arabic (Libya)", DisplayVariant: "", ISO3Country: "LBY", ISO3Language: "ara", Language: "ar", Variant: ""},
	{LocaleString: "ar-MA", Country: "MA", DisplayCountry: "Morocco", DisplayLanguage: "Arabic", DisplayName: "Arabic (Morocco)", DisplayVariant: "", ISO3Country: "MAR", ISO3Language: "ara", Language: "ar", Variant: ""},
	{LocaleString: "ar-OM", Country: "OM", DisplayCountry: "Oman", DisplayLanguage: "Arabic", DisplayName: "Arabic (Oman)", DisplayVariant: "", ISO3Country: "OMN", ISO3Language: "ara", Language: "ar", Variant: ""},
	{LocaleString: "ar-QA", Country: "QA", DisplayCountry: "Qatar", DisplayLanguage: "Arabic", DisplayName: "Arabic (Qatar)", DisplayVariant: "", ISO3Country: "QAT", ISO3Language: "ara", Language: "ar", Variant: ""},
	{LocaleString: "ar-SA", Country: "SA", DisplayCountry: "Saudi Arabia", DisplayLanguage: "Arabic", DisplayName: "Arabic (Saudi Arabia)", DisplayVariant: "", ISO3Country: "SAU", ISO3Language: "ara", Language: "ar", Variant: ""},
	{LocaleString: "ar-SD", Country: "SD", DisplayCountry: "Sudan", DisplayLanguage: "Arabic", DisplayName: "Arabic (Sudan)", DisplayVariant: "", ISO3Country: "SDN", ISO3Language: "ara", Language: "ar", Variant: ""},
	{LocaleString: "ar-SY", Country: "SY", DisplayCountry: "Syria", DisplayLanguage: "Arabic", DisplayName: "Arabic (Syria)", DisplayVariant: "", ISO3Country: "SYR", ISO3Language: "ara", Language: "ar", Variant: ""},
	{LocaleString: "ar-TN", Country: "TN", DisplayCountry: "Tunisia", DisplayLanguage: "Arabic", DisplayName: "Arabic (Tunisia)", DisplayVariant: "", ISO3Country: "TUN", ISO3Language: "ara", Language: "ar", Variant: ""},
	{LocaleString: "ar-YE", Country: "YE", DisplayCountry: "Yemen", DisplayLanguage: "Arabic", DisplayName: "Arabic (Yemen)", DisplayVariant: "", ISO3Country: "YEM", ISO3Language: "ara", Language: "ar", Variant: ""},
	{LocaleString: "be", Country: "", DisplayCountry: "", DisplayLanguage: "Belarusian", DisplayName: "Belarusian", DisplayVariant: "", ISO3Country: "", ISO3Language: "bel", Language: "be", Variant: ""},
	{LocaleString: "be-BY", Country: "BY", DisplayCountry: "Belarus", DisplayLanguage: "Belarusian", DisplayName: "Belarusian (Belarus)", DisplayVariant: "", ISO3Country: "BLR", ISO3Language: "bel", Language: "be", Variant: ""},
	{LocaleString: "bg", Country: "", DisplayCountry: "", DisplayLanguage: "Bulgarian", DisplayName: "Bulgarian", DisplayVariant: "", ISO3Country: "", ISO3Language: "bul", Language: "bg", Variant: ""},
	{LocaleString: "bg-BG", Country: "BG", DisplayCountry: "Bulgaria", DisplayLanguage: "Bulgarian", DisplayName: "Bulgarian (Bulgaria)", DisplayVariant: "", ISO3Country: "BGR", ISO3Language: "bul", Language: "bg", Variant: ""},
	{LocaleString: "ca", Country: "", DisplayCountry: "", DisplayLanguage: "Catalan", DisplayName: "Catalan", DisplayVariant: "", ISO3Country: "", ISO3Language: "cat", Language: "ca", Variant: ""},
	{LocaleString: "ca-ES", Country: "ES", DisplayCountry: "Spain", DisplayLanguage: "Catalan", DisplayName: "Catalan (Spain)", DisplayVariant: "", ISO3Country: "ESP", ISO3Language: "cat", Language: "ca", Variant: ""},
	{LocaleString: "cs", Country: "", DisplayCountry: "", DisplayLanguage: "Czech", DisplayName: "Czech", DisplayVariant: "", ISO3Country: "", ISO3Language: "ces", Language: "cs", Variant: ""},
	{LocaleString: "cs-CZ", Country: "CZ", DisplayCountry: "Czech Republic", DisplayLanguage: "Czech", DisplayName: "Czech (Czech Republic)", DisplayVariant: "", ISO3Country: "CZE", ISO3Language: "ces", Language: "cs", Variant: ""},
	{LocaleString: "da", Country: "", DisplayCountry: "", DisplayLanguage: "Danish", DisplayName: "Danish", DisplayVariant: "", ISO3Country: "", ISO3Language: "dan", Language: "da", Variant: ""},
	{LocaleString: "da-DK", Country: "DK", DisplayCountry: "Denmark", DisplayLanguage: "Danish", DisplayName: "Danish (Denmark)", DisplayVariant: "", ISO3Country: "DNK", ISO3Language: "dan", Language: "da", Variant: ""},
	{LocaleString: "de", Country: "", DisplayCountry: "", DisplayLanguage: "German", DisplayName: "German", DisplayVariant: "", ISO3Country: "", ISO3Language: "deu", Language: "de", Variant: ""},
	{LocaleString: "de-AT", Country: "AT", DisplayCountry: "Austria", DisplayLanguage: "German", DisplayName: "German (Austria)", DisplayVariant: "", ISO3Country: "AUT", ISO3Language: "deu", Language: "de", Variant: ""},
	{LocaleString: "de-BE", Country: "BE", DisplayCountry: "Belgium", DisplayLanguage: "German", DisplayName: "German (Belgium)", DisplayVariant: "", ISO3Country: "BEL", ISO3Language: "deu", Language: "de", Variant: ""},
	{LocaleString: "de-CH", Country: "CH", DisplayCountry: "Switzerland", DisplayLanguage: "German", DisplayName: "German (Switzerland)", DisplayVariant: "", ISO3Country: "CHE", ISO3Language: "deu", Language: "de", Variant: ""},
	{LocaleString: "de-DE", Country: "DE", DisplayCountry: "Germany", DisplayLanguage: "German", DisplayName: "German (Germany)", DisplayVariant: "", ISO3Country: "DEU", ISO3Language: "deu", Language: "de", Variant: ""},
	{LocaleString: "de-LU", Country: "LU", DisplayCountry: "Luxembourg", DisplayLanguage: "German", DisplayName: "German (Luxembourg)", DisplayVariant: "", ISO3Country: "LUX", ISO3Language: "deu", Language: "de", Variant: ""},
	{LocaleString: "el", Country: "", DisplayCountry: "", DisplayLanguage: "Greek", DisplayName: "Greek", DisplayVariant: "", ISO3Country: "", ISO3Language: "ell", Language: "el", Variant: ""},
	{LocaleString: "el-GR", Country: "GR", DisplayCountry: "Greece", DisplayLanguage: "Greek", DisplayName: "Greek (Greece)", DisplayVariant: "", ISO3Country: "GRC", ISO3Language: "ell", Language: "el", Variant: ""},
	{LocaleString: "en", Country: "", DisplayCountry: "", DisplayLanguage: "English", DisplayName: "English", DisplayVariant: "", ISO3Country: "", ISO3Language: "eng", Language: "en", Variant: ""},
	{LocaleString: "en-AU", Country: "AU", DisplayCountry: "Australia", DisplayLanguage: "English", DisplayName: "English (Australia)", DisplayVariant: "", ISO3Country: "AUS", ISO3Language: "eng", Language: "en", Variant: ""},
	{LocaleString: "en-CA", Country: "CA", DisplayCountry: "Canada", DisplayLanguage: "English", DisplayName: "English (Canada)", DisplayVariant: "", ISO3Country: "CAN", ISO3Language: "eng", Language: "en", Variant: ""},
	{LocaleString: "en-GB", Country: "GB", DisplayCountry: "United Kingdom", DisplayLanguage: "English", DisplayName: "English (United Kingdom)", DisplayVariant: "", ISO3Country: "GBR", ISO3Language: "eng", Language: "en", Variant: ""},
	{LocaleString: "en-IE", Country: "IE", DisplayCountry: "Ireland", DisplayLanguage: "English", DisplayName: "English (Ireland)", DisplayVariant: "", ISO3Country: "IRL", ISO3Language: "eng", Language: "en", Variant: ""},
	{LocaleString: "en-IN", Country: "IN", DisplayCountry: "India", DisplayLanguage: "English", DisplayName: "English (India)", DisplayVariant: "", ISO3Country: "IND", ISO3Language: "eng", Language: "en", Variant: ""},
	{LocaleString: "en-NZ", Country: "NZ", DisplayCountry: "New Zealand", DisplayLanguage: "English", DisplayName: "English (New Zealand)", DisplayVariant: "", ISO3Country: "NZL", ISO3Language: "eng", Language: "en", Variant: ""},
	{LocaleString: "en-US", Country: "US", DisplayCountry: "United States", DisplayLanguage: "English", DisplayName: "English (United States)", DisplayVariant: "", ISO3Country: "USA", ISO3Language: "eng", Language: "en", Variant: ""},
	{LocaleString: "en-ZA", Country: "ZA", DisplayCountry: "South Africa", DisplayLanguage: "English", DisplayName: "English (South Africa)", DisplayVariant: "", ISO3Country: "ZAF", ISO3Language: "eng", Language: "en", Variant: ""},
	{LocaleString: "es", Country: "", DisplayCountry: "", DisplayLanguage: "Spanish", DisplayName: "Spanish", DisplayVariant: "", ISO3Country: "", ISO3Language: "spa", Language: "es", Variant: ""},
	{LocaleString: "es-AR", Country: "AR", DisplayCountry: "Argentina", DisplayLanguage: "Spanish", DisplayName: "Spanish (Argentina)", DisplayVariant: "", ISO3Country: "ARG", ISO3Language: "spa", Language: "es", Variant: ""},
	{LocaleString: "es-BO", Country: "BO", DisplayCountry: "Bolivia", DisplayLanguage: "Spanish", DisplayName: "Spanish (Bolivia)", DisplayVariant: "", ISO3Country: "BOL", ISO3Language: "spa", Language: "es", Variant: ""},
	{LocaleString: "es-CL", Country: "CL", DisplayCountry: "Chile", DisplayLanguage: "Spanish", DisplayName: "Spanish (Chile)", DisplayVariant: "", ISO3Country: "CHL", ISO3Language: "spa", Language: "es", Variant: ""},
	{LocaleString: "es-CO", Country: "CO", DisplayCountry: "Colombia", DisplayLanguage: "Spanish", DisplayName: "Spanish (Colombia)", DisplayVariant: "", ISO3Country: "COL", ISO3Language: "spa", Language: "es", Variant: ""},
	{LocaleString: "es-CR", Country: "CR", DisplayCountry: "Costa Rica", DisplayLanguage: "Spanish", DisplayName: "Spanish (Costa Rica)", DisplayVariant: "", ISO3Country: "CRI", ISO3Language: "spa", Language: "es", Variant: ""},
	{LocaleString: "es-DO", Country: "DO", DisplayCountry: "Dominican Republic", DisplayLanguage: "Spanish", DisplayName: "Spanish (Dominican Republic)", DisplayVariant: "", ISO3Country: "DOM", ISO3Language: "spa", Language: "es", Variant: ""},
	{LocaleString: "es-EC", Country: "EC", DisplayCountry: "Ecuador", DisplayLanguage: "Spanish", DisplayName: "Spanish (Ecuador)", DisplayVariant: "", ISO3Country: "ECU", ISO3Language: "spa", Language: "es", Variant: ""},
	{LocaleString: "es-ES", Country: "ES", DisplayCountry: "Spain", DisplayLanguage: "Spanish", DisplayName: "Spanish (Spain)", DisplayVariant: "", ISO3Country: "ESP", ISO3Language: "spa", Language: "es", Variant: ""},
	{LocaleString: "es-GT", Country: "GT", DisplayCountry: "Guatemala", DisplayLanguage: "Spanish", DisplayName: "Spanish (Guatemala)", DisplayVariant: "", ISO3Country: "GTM", ISO3Language: "spa", Language: "es", Variant: ""},
	{LocaleString: "es-HN", Country: "HN", DisplayCountry: "Honduras", DisplayLanguage: "Spanish", DisplayName: "Spanish (Honduras)", DisplayVariant: "", ISO3Country: "HND", ISO3Language: "spa", Language: "es", Variant: ""},
	{LocaleString: "es-MX", Country: "MX", DisplayCountry: "Mexico", DisplayLanguage: "Spanish", DisplayName: "Spanish (Mexico)", DisplayVariant: "", ISO3Country: "MEX", ISO3Language: "spa", Language: "es", Variant: ""},
	{LocaleString: "es-NI", Country: "NI", DisplayCountry: "Nicaragua", DisplayLanguage: "Spanish", DisplayName: "Spanish (Nicaragua)", DisplayVariant: "", ISO3Country: "NIC", ISO3Language: "spa", Language: "es", Variant: ""},
	{LocaleString: "es-PA", Country: "PA", DisplayCountry: "Panama", DisplayLanguage: "Spanish", DisplayName: "Spanish (Panama)", DisplayVariant: "", ISO3Country: "PAN", ISO3Language: "spa", Language: "es", Variant: ""},
	{LocaleString: "es-PE", Country: "PE", DisplayCountry: "Peru", DisplayLanguage: "Spanish", DisplayName: "Spanish (Peru)", DisplayVariant: "", ISO3Country: "PER", ISO3Language: "spa", Language: "es", Variant: ""},
	{LocaleString: "es-PR", Country: "PR", DisplayCountry: "Puerto Rico", DisplayLanguage: "Spanish", DisplayName: "Spanish (Puerto Rico)", DisplayVariant: "", ISO3Country: "PRI", ISO3Language: "spa", Language: "es", Variant: ""},
	{LocaleString: "es-PY", Country: "PY", DisplayCountry: "Paraguay", DisplayLanguage: "Spanish", DisplayName: "Spanish (Paraguay)", DisplayVariant: "", ISO3Country: "PRY", ISO3Language: "spa", Language: "es", Variant: ""},
	{LocaleString: "es-SV", Country: "SV", DisplayCountry: "El Salvador", DisplayLanguage: "Spanish", DisplayName: "Spanish (El Salvador)", DisplayVariant: "", ISO3Country: "SLV", ISO3Language: "spa", Language: "es", Variant: ""},
	{LocaleString: "es-UY", Country: "UY", DisplayCountry: "Uruguay", DisplayLanguage: "Spanish", DisplayName: "Spanish (Uruguay)", DisplayVariant: "", ISO3Country: "URY", ISO3Language: "spa", Language: "es", Variant: ""},
	{LocaleString: "es-VE", Country: "VE", DisplayCountry: "Venezuela", DisplayLanguage: "Spanish", DisplayName: "Spanish (Venezuela)", DisplayVariant: "", ISO3Country: "VEN", ISO3Language: "spa", Language: "es", Variant: ""},
	{LocaleString: "et", Country: "", DisplayCountry: "", DisplayLanguage: "Estonian", DisplayName: "Estonian", DisplayVariant: "", ISO3Country: "", ISO3Language: "est", Language: "et", Variant: ""},
	{LocaleString: "et-EE", Country: "EE", DisplayCountry: "Estonia", DisplayLanguage: "Estonian", DisplayName: "Estonian (Estonia)", DisplayVariant: "", ISO3Country: "EST", ISO3Language: "est", Language: "et", Variant: ""},
	{LocaleString: "fi", Country: "", DisplayCountry: "", DisplayLanguage: "Finnish", DisplayName: "Finnish", DisplayVariant: "", ISO3Country: "", ISO3Language: "fin", Language: "fi", Variant: ""},
	{LocaleString: "fi-FI", Country: "FI", DisplayCountry: "Finland", DisplayLanguage: "Finnish", DisplayName: "Finnish (Finland)", DisplayVariant: "", ISO3Country: "FIN", ISO3Language: "fin", Language: "fi", Variant: ""},
	{LocaleString: "fr", Country: "", DisplayCountry: "", DisplayLanguage: "French", DisplayName: "French", DisplayVariant: "", ISO3Country: "", ISO3Language: "fra", Language: "fr", Variant: ""},
	{LocaleString: "fr-BE", Country: "BE", DisplayCountry: "Belgium", DisplayLanguage: "French", DisplayName: "French (Belgium)", DisplayVariant: "", ISO3Country: "BEL", ISO3Language: "fra", Language: "fr", Variant: ""},
	{LocaleString: "fr-CA", Country: "CA", DisplayCountry: "Canada", DisplayLanguage: "French", DisplayName: "French (Canada)", DisplayVariant: "", ISO3Country: "CAN", ISO3Language: "fra", Language: "fr", Variant: ""},
	{LocaleString: "fr-CH", Country: "CH", DisplayCountry: "Switzerland", DisplayLanguage: "French", DisplayName: "French (Switzerland)", DisplayVariant: "", ISO3Country: "CHE", ISO3Language: "fra", Language: "fr", Variant: ""},
	{LocaleString: "fr-FR", Country: "FR", DisplayCountry: "France", DisplayLanguage: "French", DisplayName: "French (France)", DisplayVariant: "", ISO3Country: "FRA", ISO3Language: "fra", Language: "fr", Variant: ""},
	{LocaleString: "fr-LU", Country: "LU", DisplayCountry: "Luxembourg", DisplayLanguage: "French", DisplayName: "French (Luxembourg)", DisplayVariant: "", ISO3Country: "LUX", ISO3Language: "fra", Language: "fr", Variant: ""},
	{LocaleString: "hi-IN", Country: "IN", DisplayCountry: "India", DisplayLanguage: "Hindi", DisplayName: "Hindi (India)", DisplayVariant: "", ISO3Country: "IND", ISO3Language: "hin", Language: "hi", Variant: ""},
	{LocaleString: "hr", Country: "", DisplayCountry: "", DisplayLanguage: "Croatian", DisplayName: "Croatian", DisplayVariant: "", ISO3Country: "", ISO3Language: "hrv", Language: "hr", Variant: ""},
	{LocaleString: "hr-HR", Country: "HR", DisplayCountry: "Croatia", DisplayLanguage: "Croatian", DisplayName: "Croatian (Croatia)", DisplayVariant: "", ISO3Country: "HRV", ISO3Language: "hrv", Language: "hr", Variant: ""},
	{LocaleString: "hu", Country: "", DisplayCountry: "", DisplayLanguage: "Hungarian", DisplayName: "Hungarian", DisplayVariant: "", ISO3Country: "", ISO3Language: "hun", Language: "hu", Variant: ""},
	{LocaleString: "hu-HU", Country: "HU", DisplayCountry: "Hungary", DisplayLanguage: "Hungarian", DisplayName: "Hungarian (Hungary)", DisplayVariant: "", ISO3Country: "HUN", ISO3Language: "hun", Language: "hu", Variant: ""},
	{LocaleString: "is", Country: "", DisplayCountry: "", DisplayLanguage: "Icelandic", DisplayName: "Icelandic", DisplayVariant: "", ISO3Country: "", ISO3Language: "isl", Language: "is", Variant: ""},
	{LocaleString: "is-IS", Country: "IS", DisplayCountry: "Iceland", DisplayLanguage: "Icelandic", DisplayName: "Icelandic (Iceland)", DisplayVariant: "", ISO3Country: "ISL", ISO3Language: "isl", Language: "is", Variant: ""},
	{LocaleString: "it", Country: "", DisplayCountry: "", DisplayLanguage: "Italian", DisplayName: "Italian", DisplayVariant: "", ISO3Country: "", ISO3Language: "ita", Language: "it", Variant: ""},
	{LocaleString: "it-CH", Country: "CH", DisplayCountry: "Switzerland", DisplayLanguage: "Italian", DisplayName: "Italian (Switzerland)", DisplayVariant: "", ISO3Country: "CHE", ISO3Language: "ita", Language: "it", Variant: ""},
	{LocaleString: "it-IT", Country: "IT", DisplayCountry: "Italy", DisplayLanguage: "Italian", DisplayName: "Italian (Italy)", DisplayVariant: "", ISO3Country: "ITA", ISO3Language: "ita", Language: "it", Variant: ""},
	{LocaleString: "iw", Country: "", DisplayCountry: "", DisplayLanguage: "Hebrew", DisplayName: "Hebrew", DisplayVariant: "", ISO3Country: "", ISO3Language: "heb", Language: "iw", Variant: ""},
	{LocaleString: "iw-IL", Country: "IL", DisplayCountry: "Israel", DisplayLanguage: "Hebrew", DisplayName: "Hebrew (Israel)", DisplayVariant: "", ISO3Country: "ISR", ISO3Language: "heb", Language: "iw", Variant: ""},
	{LocaleString: "ja", Country: "", DisplayCountry: "", DisplayLanguage: "Japanese", DisplayName: "Japanese", DisplayVariant: "", ISO3Country: "", ISO3Language: "jpn", Language: "ja", Variant: ""},
	{LocaleString: "ja-JP", Country: "JP", DisplayCountry: "Japan", DisplayLanguage: "Japanese", DisplayName: "Japanese (Japan)", DisplayVariant: "", ISO3Country: "JPN", ISO3Language: "jpn", Language: "ja", Variant: ""},
	{LocaleString: "ko", Country: "", DisplayCountry: "", DisplayLanguage: "Korean", DisplayName: "Korean", DisplayVariant: "", ISO3Country: "", ISO3Language: "kor", Language: "ko", Variant: ""},
	{LocaleString: "ko-KR", Country: "KR", DisplayCountry: "South Korea", DisplayLanguage: "Korean", DisplayName: "Korean (South Korea)", DisplayVariant: "", ISO3Country: "KOR", ISO3Language: "kor", Language: "ko", Variant: ""},
	{LocaleString: "lt", Country: "", DisplayCountry: "", DisplayLanguage: "Lithuanian", DisplayName: "Lithuanian", DisplayVariant: "", ISO3Country: "", ISO3Language: "lit", Language: "lt", Variant: ""},
	{LocaleString: "lt-LT", Country: "LT", DisplayCountry: "Lithuania", DisplayLanguage: "Lithuanian", DisplayName: "Lithuanian (Lithuania)", DisplayVariant: "", ISO3Country: "LTU", ISO3Language: "lit", Language: "lt", Variant: ""},
	{LocaleString: "lv", Country: "", DisplayCountry: "", DisplayLanguage: "Latvian", DisplayName: "Latvian", DisplayVariant: "", ISO3Country: "", ISO3Language: "lav", Language: "lv", Variant: ""},
	{LocaleString: "lv-LV", Country: "LV", DisplayCountry: "Latvia", DisplayLanguage: "Latvian", DisplayName: "Latvian (Latvia)", DisplayVariant: "", ISO3Country: "LVA", ISO3Language: "lav", Language: "lv", Variant: ""},
	{LocaleString: "mk", Country: "", DisplayCountry: "", DisplayLanguage: "Macedonian", DisplayName: "Macedonian", DisplayVariant: "", ISO3Country: "", ISO3Language: "mkd", Language: "mk", Variant: ""},
	{LocaleString: "mk-MK", Country: "MK", DisplayCountry: "Macedonia", DisplayLanguage: "Macedonian", DisplayName: "Macedonian (Macedonia)", DisplayVariant: "", ISO3Country: "MKD", ISO3Language: "mkd", Language: "mk", Variant: ""},
	{LocaleString: "nl", Country: "", DisplayCountry: "", DisplayLanguage: "Dutch", DisplayName: "Dutch", DisplayVariant: "", ISO3Country: "", ISO3Language: "nld", Language: "nl", Variant: ""},
	{LocaleString: "nl-BE", Country: "BE", DisplayCountry: "Belgium", DisplayLanguage: "Dutch", DisplayName: "Dutch (Belgium)", DisplayVariant: "", ISO3Country: "BEL", ISO3Language: "nld", Language: "nl", Variant: ""},
	{LocaleString: "nl-NL", Country: "NL", DisplayCountry: "Netherlands", DisplayLanguage: "Dutch", DisplayName: "Dutch (Netherlands)", DisplayVariant: "", ISO3Country: "NLD", ISO3Language: "nld", Language: "nl", Variant: ""},
	{LocaleString: "no", Country: "", DisplayCountry: "", DisplayLanguage: "Norwegian", DisplayName: "Norwegian", DisplayVariant: "", ISO3Country: "", ISO3Language: "nor", Language: "no", Variant: ""},
	{LocaleString: "no-NO", Country: "NO", DisplayCountry: "Norway", DisplayLanguage: "Norwegian", DisplayName: "Norwegian (Norway)", DisplayVariant: "", ISO3Country: "NOR", ISO3Language: "nor", Language: "no", Variant: ""},
	{LocaleString: "no-NO-NY", Country: "NO", DisplayCountry: "Norway", DisplayLanguage: "Norwegian", DisplayName: "Norwegian (Norway,Nynorsk)", DisplayVariant: "Nynorsk", ISO3Country: "NOR", ISO3Language: "nor", Language: "no", Variant: "NY"},
	{LocaleString: "pl", Country: "", DisplayCountry: "", DisplayLanguage: "Polish", DisplayName: "Polish", DisplayVariant: "", ISO3Country: "", ISO3Language: "pol", Language: "pl", Variant: ""},
	{LocaleString: "pl-PL", Country: "PL", DisplayCountry: "Poland", DisplayLanguage: "Polish", DisplayName: "Polish (Poland)", DisplayVariant: "", ISO3Country: "POL", ISO3Language: "pol", Language: "pl", Variant: ""},
	{LocaleString: "pt", Country: "", DisplayCountry: "", DisplayLanguage: "Portuguese", DisplayName: "Portuguese", DisplayVariant: "", ISO3Country: "", ISO3Language: "por", Language: "pt", Variant: ""},
	{LocaleString: "pt-BR", Country: "BR", DisplayCountry: "Brazil", DisplayLanguage: "Portuguese", DisplayName: "Portuguese (Brazil)", DisplayVariant: "", ISO3Country: "BRA", ISO3Language: "por", Language: "pt", Variant: ""},
	{LocaleString: "pt-PT", Country: "PT", DisplayCountry: "Portugal", DisplayLanguage: "Portuguese", DisplayName: "Portuguese (Portugal)", DisplayVariant: "", ISO3Country: "PRT", ISO3Language: "por", Language: "pt", Variant: ""},
	{LocaleString: "ro", Country: "", DisplayCountry: "", DisplayLanguage: "Romanian", DisplayName: "Romanian", DisplayVariant: "", ISO3Country: "", ISO3Language: "ron", Language: "ro", Variant: ""},
	{LocaleString: "ro-RO", Country: "RO", DisplayCountry: "Romania", DisplayLanguage: "Romanian", DisplayName: "Romanian (Romania)", DisplayVariant: "", ISO3Country: "ROU", ISO3Language: "ron", Language: "ro", Variant: ""},
	{LocaleString: "ru", Country: "", DisplayCountry: "", DisplayLanguage: "Russian", DisplayName: "Russian", DisplayVariant: "", ISO3Country: "", ISO3Language: "rus", Language: "ru", Variant: ""},
	{LocaleString: "ru-RU", Country: "RU", DisplayCountry: "Russia", DisplayLanguage: "Russian", DisplayName: "Russian (Russia)", DisplayVariant: "", ISO3Country: "RUS", ISO3Language: "rus", Language: "ru", Variant: ""},
	{LocaleString: "sk", Country: "", DisplayCountry: "", DisplayLanguage: "Slovak", DisplayName: "Slovak", DisplayVariant: "", ISO3Country: "", ISO3Language: "slk", Language: "sk", Variant: ""},
	{LocaleString: "sk-SK", Country: "SK", DisplayCountry: "Slovakia", DisplayLanguage: "Slovak", DisplayName: "Slovak (Slovakia)", DisplayVariant: "", ISO3Country: "SVK", ISO3Language: "slk", Language: "sk", Variant: ""},
	{LocaleString: "sl", Country: "", DisplayCountry: "", DisplayLanguage: "Slovenian", DisplayName: "Slovenian", DisplayVariant: "", ISO3Country: "", ISO3Language: "slv", Language: "sl", Variant: ""},
	{LocaleString: "sl-SI", Country: "SI", DisplayCountry: "Slovenia", DisplayLanguage: "Slovenian", DisplayName: "Slovenian (Slovenia)", DisplayVariant: "", ISO3Country: "SVN", ISO3Language: "slv", Language: "sl", Variant: ""},
	{LocaleString: "sq", Country: "", DisplayCountry: "", DisplayLanguage: "Albanian", DisplayName: "Albanian", DisplayVariant: "", ISO3Country: "", ISO3Language: "sqi", Language: "sq", Variant: ""},
	{LocaleString: "sq-AL", Country: "AL", DisplayCountry: "Albania", DisplayLanguage: "Albanian", DisplayName: "Albanian (Albania)", DisplayVariant: "", ISO3Country: "ALB", ISO3Language: "sqi", Language: "sq", Variant: ""},
	{LocaleString: "sr", Country: "", DisplayCountry: "", DisplayLanguage: "Serbian", DisplayName: "Serbian", DisplayVariant: "", ISO3Country: "", ISO3Language: "srp", Language: "sr", Variant: ""},
	{LocaleString: "sr-BA", Country: "BA", DisplayCountry: "Bosnia and Herzegovina", DisplayLanguage: "Serbian", DisplayName: "Serbian (Bosnia and Herzegovina)", DisplayVariant: "", ISO3Country: "BIH", ISO3Language: "srp", Language: "sr", Variant: ""},
	{LocaleString: "sr-CS", Country: "CS", DisplayCountry: "Serbia and Montenegro", DisplayLanguage: "Serbian", DisplayName: "Serbian (Serbia and Montenegro)", DisplayVariant: "", ISO3Country: "SCG", ISO3Language: "srp", Language: "sr", Variant: ""},
	{LocaleString: "sv", Country: "", DisplayCountry: "", DisplayLanguage: "Swedish", DisplayName: "Swedish", DisplayVariant: "", ISO3Country: "", ISO3Language: "swe", Language: "sv", Variant: ""},
	{LocaleString: "sv-SE", Country: "SE", DisplayCountry: "Sweden", DisplayLanguage: "Swedish", DisplayName: "Swedish (Sweden)", DisplayVariant: "", ISO3Country: "SWE", ISO3Language: "swe", Language: "sv", Variant: ""},
	{LocaleString: "th", Country: "", DisplayCountry: "", DisplayLanguage: "Thai", DisplayName: "Thai", DisplayVariant: "", ISO3Country: "", ISO3Language: "tha", Language: "th", Variant: ""},
	{LocaleString: "th-TH", Country: "TH", DisplayCountry: "Thailand", DisplayLanguage: "Thai", DisplayName: "Thai (Thailand)", DisplayVariant: "", ISO3Country: "THA", ISO3Language: "tha", Language: "th", Variant: ""},
	{LocaleString: "th-TH-TH", Country: "TH", DisplayCountry: "Thailand", DisplayLanguage: "Thai", DisplayName: "Thai (Thailand,TH)", DisplayVariant: "TH", ISO3Country: "THA", ISO3Language: "tha", Language: "th", Variant: "TH"},
	{LocaleString: "tr", Country: "", DisplayCountry: "", DisplayLanguage: "Turkish", DisplayName: "Turkish", DisplayVariant: "", ISO3Country: "", ISO3Language: "tur", Language: "tr", Variant: ""},
	{LocaleString: "tr-TR", Country: "TR", DisplayCountry: "Turkey", DisplayLanguage: "Turkish", DisplayName: "Turkish (Turkey)", DisplayVariant: "", ISO3Country: "TUR", ISO3Language: "tur", Language: "tr", Variant: ""},
	{LocaleString: "uk", Country: "", DisplayCountry: "", DisplayLanguage: "Ukrainian", DisplayName: "Ukrainian", DisplayVariant: "", ISO3Country: "", ISO3Language: "ukr", Language: "uk", Variant: ""},
	{LocaleString: "uk-UA", Country: "UA", DisplayCountry: "Ukraine", DisplayLanguage: "Ukrainian", DisplayName: "Ukrainian (Ukraine)", DisplayVariant: "", ISO3Country: "UKR", ISO3Language: "ukr", Language: "uk", Variant: ""},
	{LocaleString: "vi", Country: "", DisplayCountry: "", DisplayLanguage: "Vietnamese", DisplayName: "Vietnamese", DisplayVariant: "", ISO3Country: "", ISO3Language: "vie", Language: "vi", Variant: ""},
	{LocaleString: "vi-VN", Country: "VN", DisplayCountry: "Vietnam", DisplayLanguage: "Vietnamese", DisplayName: "Vietnamese (Vietnam)", DisplayVariant: "", ISO3Country: "VNM", ISO3Language: "vie", Language: "vi", Variant: ""},
	{LocaleString: "zh", Country: "", DisplayCountry: "", DisplayLanguage: "Chinese", DisplayName: "Chinese", DisplayVariant: "", ISO3Country: "", ISO3Language: "zho", Language: "zh", Variant: ""},
	{LocaleString: "zh-CN", Country: "CN", DisplayCountry: "China", DisplayLanguage: "Chinese", DisplayName: "Chinese (China)", DisplayVariant: "", ISO3Country: "CHN", ISO3Language: "zho", Language: "zh", Variant: ""},
	{LocaleString: "zh-HK", Country: "HK", DisplayCountry: "Hong Kong", DisplayLanguage: "Chinese", DisplayName: "Chinese (Hong Kong)", DisplayVariant: "", ISO3Country: "HKG", ISO3Language: "zho", Language: "zh", Variant: ""},
	{LocaleString: "zh-TW", Country: "TW", DisplayCountry: "Taiwan", DisplayLanguage: "Chinese", DisplayName: "Chinese (Taiwan)", DisplayVariant: "", ISO3Country: "TWN", ISO3Language: "zho", Language: "zh", Variant: ""},
	{LocaleString: "zh-Hans", Country: "CN", DisplayCountry: "China", DisplayLanguage: "Simplified Chinese", DisplayName: "Chinese (Simplified, China)", DisplayVariant: "", ISO3Country: "CHN", ISO3Language: "zho", Language: "zh", Variant: "Hans"},
	{LocaleString: "zh-Hant", Country: "TW", DisplayCountry: "Taiwan", DisplayLanguage: "Traditional Chinese", DisplayName: "Chinese (Traditional, Taiwan)", DisplayVariant: "", ISO3Country: "TWN", ISO3Language: "zho", Language: "zh", Variant: "Hant"},
	{LocaleString: "rm", Country: "", DisplayCountry: "", DisplayLanguage: "Romansh", DisplayName: "Romansh", DisplayVariant: "", ISO3Country: "", ISO3Language: "roh", Language: "rm", Variant: ""},
	{LocaleString: "rm-CH", Country: "CH", DisplayCountry: "Switzerland", DisplayLanguage: "Romansh", DisplayName: "Romansh (Switzerland)", DisplayVariant: "", ISO3Country: "CHE", ISO3Language: "roh", Language: "rm", Variant: ""},
	{LocaleString: "id", Country: "ID", DisplayCountry: "Indonesia", DisplayLanguage: "Indonesian", DisplayName: "Indonesian (Indonesia)", DisplayVariant: "", ISO3Country: "IDN", ISO3Language: "ind", Language: "id", Variant: ""},
}

var localeToLanguageMap map[string]string

// AllowedBCP47Locales is the list of known locale tags (BCP47: language subtag
// lower case, region/script upper when present, e.g. en-US, de-DE, zh-Hant; or
// a language-only tag like en). Matches LocaleString values in [Locales].
var AllowedBCP47Locales []string

// init initializes the locale lookup map for efficient language code retrieval.
// This function runs automatically when the package is imported.
func init() {
	localeToLanguageMap = make(map[string]string, len(Locales))
	AllowedBCP47Locales = make([]string, len(Locales))
	for i, locale := range Locales {
		AllowedBCP47Locales[i] = locale.LocaleString
		localeToLanguageMap[locale.LocaleString] = locale.Language
	}
}

// GetLanguageForLocale returns the language code for a given locale string.
// Returns "en" as default if the locale is not found.
func GetLanguageForLocale(locale string) string {
	if lang, exists := localeToLanguageMap[locale]; exists {
		return lang
	}
	return DefaultLanguage
}

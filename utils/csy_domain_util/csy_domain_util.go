package csy_domain_util

import (
	"regexp"
	"strings"
)

// 目前市面上所有的域名后缀 数据来自于 狗爹
var domain_last_pre []string = []string{
	"cashbackbonus", "international", "lifeinsurance", "spreadbetting", "construction", "scholarships", "translations", "financialaid", "productions", "investments", "enterprises", "photography", "mutualfunds", "motorcycles", "persiangulf", "accountants", "engineering", "contractors", "consulting", "Protection", "properties", "basketball", "realestate", "accountant", "republican", "retirement", "immobilien", "apartments", "industries", "university", "restaurant", "associates", "creditcard", "technology", "healthcare", "vlaanderen", "management", "foundation", "stockholm", "equipment", "analytics", "community", "barcelona", "education", "solutions", "architect", "furniture", "lifestyle", "marketing", "financial", "vacations", "aquitaine", "institute", "insurance", "amsterdam", "directory", "melbourne", "capetown", "cityeats", "movistar", "computer", "supplies", "baseball", "exchange", "download", "shopping", "builders", "brussels", "yokohama", "pharmacy", "feedback", "clothing", "cleaning", "airforce", "abudhabi", "memorial", "football", "saarland", "helsinki", "partners", "security", "broadway", "software", "attorney", "catholic", "engineer", "lighting", "business", "mortgage", "catering", "training", "ventures", "bargains", "services", "budapest", "pictures", "holdings", "boutique", "istanbul", "delivery", "discount", "diamonds", "plumbing", "hospital", "graphics", "democrat", "surgery", "trading", "network", "limited", "finance", "company", "contact", "corsica", "fashion", "auction", "info.pl", "jewelry", "support", "watches", "shiksha", "audible", "holiday", "florist", "country", "organic", "firm.in", "digital", "zuerich", "kitchen", "website", "monster", "cooking", "domains", "cruises", "grocery", "hamburg", "reviews", "hoteles", "gallery", "charity", "capital", "forsale", "ismaili", "flights", "academy", "okinawa", "markets", "tickets", "singles", "college", "coupons", "fitness", "info.ve", "recipes", "cologne", "storage", "indians", "booking", "beknown", "exposed", "theater", "dentist", "fishing", "theatre", "compare", "science", "wedding", "express", "courses", "systems", "cricket", "medical", "realtor", "rentals", "wanggou", "careers", "net.nz", "org.ag", "imamat", "racing", "museum", "com.br", "durban", "org.in", "vision", "soccer", "berlin", "active", "hotels", "market", "taipei", "sports", "reisen", "dating", "madrid", "latino", "comsec", "tienda", "center", "energy", "idv.tw", "com.mx", "agency", "degree", "camera", "online", "webcam", "dental", "com.ag", "tennis", "org.pe", "credit", "moscow", "sydney", "joburg", "com.ve", "schule", "review", "cruise", "dealer", "net.ve", "secure", "family", "net.co", "hoteis", "org.es", "broker", "casino", "voyage", "living", "net.in", "doctor", "org.ph", "nom.pe", "nom.es", "org.ve", "author", "estate", "net.pe", "net.ph", "net.bz", "insure", "condos", "com.bz", "ryukyu", "banque", "report", "com.co", "studio", "bayern", "com.pl", "mobile", "kinder", "events", "nom.co", "repair", "giving", "org.ru", "org.uk", "social", "net.br", "health", "direct", "web.ve", "com.pe", "biz.pl", "net.ru", "lawyer", "org.au", "quebec", "london", "org.cn", "physio", "career", "search", "gratis", "net.au", "travel", "design", "garden", "com.ph", "coffee", "africa", "gov.cn", "nagoya", "alsace", "org.nz", "clinic", "expert", "boston", "stream", "net.ag", "com.ru", "net.cn", "org.pl", "com.tw", "global", "hockey", "supply", "ind.in", "gen.in", "com.es", "school", "church", "kaufen", "coupon", "net.pl", "mutual", "com.cn", "photos", "villas", "claims", "viajes", "luxury", "com.au", "maison", "futbol", "beauty", "money", "ninja", "nx.cn", "wales", "koeln", "irish", "loans", "ln.cn", "weibo", "fj.cn", "deals", "adult", "reise", "codes", "earth", "boats", "movie", "music", "gripe", "rodeo", "hotel", "tools", "actor", "cards", "watch", "sport", "green", "salon", "swiss", "world", "forum", "glass", "archi", "lease", "co.uk", "games", "co.kr", "mo.cn", "media", "cheap", "rocks", "space", "lotto", "re.kr", "legal", "ac.cn", "press", "hb.cn", "osaka", "co.za", "hi.cn", "tunes", "house", "guide", "gives", "trade", "coach", "cq.cn", "miami", "sucks", "tires", "solar", "dubai", "homes", "gz.cn", "trust", "halal", "jx.cn", "paris", "ne.kr", "hn.cn", "cymru", "tirol", "prime", "jetzt", "poker", "autos", "works", "js.cn", "party", "study", "vodka", "build", "tokyo", "rugby", "shoes", "me.uk", "pizza", "today", "email", "faith", "gs.cn", "promo", "bible", "black", "bj.cn", "gx.cn", "video", "bingo", "co.ve", "group", "store", "hl.cn", "parts", "vegas", "drive", "jl.cn", "phone", "tours", "ha.cn", "he.cn", "horse", "gd.cn", "kyoto", "co.nz", "radio", "islam", "dance", "cloud", "co.in", "style", "gifts", "hk.cn", "nm.cn", "ah.cn", "rehab", "vote", "gent", "cool", "voto", "bike", "love", "arte", "info", "blog", "blue", "fail", "army", "cafe", "call", "tips", "porn", "camp", "rest", "webs", "show", "club", "song", "tube", "wiki", "zone", "case", "site", "baby", "roma", "wang", "name", "fish", "deal", "scot", "ecom", "food", "wine", "free", "city", "beer", "page", "golf", "chat", "land", "life", "sale", "book", "here", "thai", "navy", "doha", "pets", "auto", "cars", "casa", "host", "best", "seat", "yoga", "gmbh", "immo", "kids", "care", "live", "rsvp", "buzz", "fund", "kiwi", "reit", "hair", "aero", "town", "surf", "film", "qpon", "news", "ltda", "guru", "save", "silk", "mail", "luxe", "coop", "menu", "moda", "lgbt", "tech", "data", "asia", "work", "farm", "taxi", "team", "arab", "docs", "loan", "prof", "room", "shia", "date", "pink", "haus", "toys", "gold", "band", "fans", "rent", "cash", "plus", "limo", "tour", "bank", "shop", "sarl", "wien", "rich", "dog", "xin", "rip", "ceo", "map", "esq", "bet", "cam", "mls", "moe", "cab", "xxx", "cpa", "app", "fan", "top", "dev", "bid", "uno", "nyc", "wed", "ren", "bot", "med", "bar", "pay", "kid", "xyz", "fit", "sex", "idv", "mil", "bio", "eat", "ngo", "rio", "nrw", "buy", "gle", "gay", "int", "Tel", "red", "pet", "wtf", "soy", "mba", "men", "new", "mov", "dad", "cfd", "zip", "biz", "diy", "ged", "net", "fun", "edu", "pub", "box", "ira", "bzh", "art", "ski", "vip", "eus", "yun", "law", "onl", "phd", "spa", "llc", "llp", "srl", "eco", "pro", "kim", "ltd", "Pro", "web", "ink", "car", "fyi", "win", "inc", "org", "dds", "ist", "com", "gov", "vin", "run", "dot", "ads", "tax", "lat", "one", "vet", "hu", "st", "fm", "np", "nl", "de", "gy", "co", "be", "sd", "lk", "sv", "wf", "yt", "cv", "bo", "eh", "bz", "ws", "at", "ie", "bd", "nf", "gt", "py", "as", "cf", "la", "bm", "gu", "gf", "es", "pe", "sg", "gw", "mw", "tr", "dk", "cc", "cn", "sy", "se", "cm", "pn", "ht", "lv", "ug", "li", "pm", "lr", "tk", "kr", "vg", "tn", "sz", "ca", "kw", "tc", "am", "ki", "sc", "bw", "bi", "mn", "gh", "ly", "lc", "ao", "pk", "km", "si", "tf", "th", "ky", "mo", "ag", "mz", "vn", "jp", "mc", "mq", "fo", "io", "kh", "dm", "zm", "gd", "mh", "mr", "za", "jm", "bj", "vu", "et", "ru", "an", "tj", "ga", "pg", "gs", "uy", "by", "bs", "lt", "gr", "va", "ye", "br", "ke", "tl", "ch", "ne", "lq", "ac", "bf", "nz", "bn", "uk", "ed", "cr", "us", "ma", "ad", "je", "qa", "cz", "il", "ec", "hk", "nc", "al", "re", "gp", "gg", "nr", "aw", "kp", "ml", "jo", "pf", "kz", "sa", "eu", "bh", "tg", "ee", "cu", "cl", "af", "um", "is", "mm", "bt", "mx", "kn", "nu", "cd", "bv", "tm", "er", "bb", "my", "na", "fi", "fr", "gi", "tw", "lb", "au", "so", "no", "lu", "ro", "sm", "sb", "pr", "sk", "mp", "fj", "vi", "tv", "in", "bg", "mt", "tt", "ba", "tz", "pa", "rw", "sr", "sl", "om", "md", "az", "ni", "mg", "ve", "fk", "gm", "to", "pl", "dj", "ph", "mu", "ir", "vc", "ps", "ls", "ge", "do", "sn", "cx", "id", "uz", "eg", "mv", "im", "yu", "dz", "sh", "ck", "hm", "pt", "cg", "kg", "ar", "pw", "ci", "mk", "cy", "aq", "hr", "zw", "gn", "ai", "gl", "me", "ms", "it",
}

func DomainRootName(domain string) (last_domain string) {
	if !regexp.MustCompile(`([a-zA-Z0-9-]+.)+([a-zA-Z])+$`).MatchString(domain) {
		return
	}

	domain = strings.ToLower(domain)
	//domain := "q2wfsg.asfsdg.fsdg.bj.cn"
	domain = regexp.MustCompile(`[ ]+`).ReplaceAllString(domain, "")
	for _, d_type := range domain_last_pre {
		if strings.HasSuffix(domain, "."+d_type) {
			//提取主域名
			tmp_domain := domain[:strings.LastIndex(domain, "."+d_type)]
			tmp_doamin_split := strings.Split(tmp_domain, ".")
			zhu_domain := tmp_doamin_split[0]
			if len(tmp_doamin_split) > 1 {
				zhu_domain = tmp_doamin_split[len(tmp_doamin_split)-1]
			}
			last_domain = zhu_domain + "." + d_type
			break
		}
	}
	return
}
func DomainRootLastFix(domain string) (last_domain string) {
	domain = strings.ToLower(domain)
	//domain := "q2wfsg.asfsdg.fsdg.bj.cn"
	domain = regexp.MustCompile(`[ ]+`).ReplaceAllString(domain, "")
	for _, d_type := range domain_last_pre {
		if strings.HasSuffix(domain, "."+d_type) {
			last_domain = d_type
			break
		}
	}
	return
}

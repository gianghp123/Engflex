package pronunciation

// MinimalPair is one contrast the learner hears, discriminates, then repeats.
// The IDs are display-only; the server never resolves them to a row.
type MinimalPair struct {
	A          string `json:"a"`
	B          string `json:"b"`
	IPAA       string `json:"ipaA"`
	IPAB       string `json:"ipaB"`
	ContrastID string `json:"contrastId"`
}

// Pairs is the static VN-learner contrast set. It is intentionally small:
// every entry maps to a coaching rule in vn_rules.go, and the web card
// renders the same list locally (no endpoint, no persistence).
var Pairs = []MinimalPair{
	{A: "think", B: "tink", IPAA: "/θɪŋk/", IPAB: "/tɪŋk/", ContrastID: "th-t"},
	{A: "thick", B: "tick", IPAA: "/θɪk/", IPAB: "/tɪk/", ContrastID: "th-t"},
	{A: "three", B: "tree", IPAA: "/θriː/", IPAB: "/triː/", ContrastID: "th-t"},
	{A: "they", B: "day", IPAA: "/ðeɪ/", IPAB: "/deɪ/", ContrastID: "th-d"},
	{A: "then", B: "den", IPAA: "/ðen/", IPAB: "/den/", ContrastID: "th-d"},
	{A: "breathe", B: "breed", IPAA: "/briːð/", IPAB: "/briːd/", ContrastID: "th-d"},
	{A: "she", B: "see", IPAA: "/ʃiː/", IPAB: "/siː/", ContrastID: "sh-s"},
	{A: "ship", B: "sip", IPAA: "/ʃɪp/", IPAB: "/sɪp/", ContrastID: "sh-s"},
	{A: "shoe", B: "Sue", IPAA: "/ʃuː/", IPAB: "/suː/", ContrastID: "sh-s"},
	{A: "chew", B: "true", IPAA: "/tʃuː/", IPAB: "/truː/", ContrastID: "ch-tr"},
	{A: "chain", B: "train", IPAA: "/tʃeɪn/", IPAB: "/treɪn/", ContrastID: "ch-tr"},
	{A: "light", B: "night", IPAA: "/laɪt/", IPAB: "/naɪt/", ContrastID: "l-n"},
	{A: "lap", B: "nap", IPAA: "/læp/", IPAB: "/næp/", ContrastID: "l-n"},
	{A: "low", B: "no", IPAA: "/ləʊ/", IPAB: "/nəʊ/", ContrastID: "l-n"},
	{A: "sing", B: "sin", IPAA: "/sɪŋ/", IPAB: "/sɪn/", ContrastID: "ng-n"},
	{A: "thing", B: "thin", IPAA: "/θɪŋ/", IPAB: "/θɪn/", ContrastID: "ng-n"},
	{A: "vest", B: "west", IPAA: "/vest/", IPAB: "/west/", ContrastID: "v-w"},
	{A: "vine", B: "wine", IPAA: "/vaɪn/", IPAB: "/waɪn/", ContrastID: "v-w"},
	{A: "zip", B: "sip", IPAA: "/zɪp/", IPAB: "/sɪp/", ContrastID: "z-s"},
	{A: "big", B: "pig", IPAA: "/bɪɡ/", IPAB: "/pɪɡ/", ContrastID: "b-p"},
	{A: "buy", B: "pie", IPAA: "/baɪ/", IPAB: "/paɪ/", ContrastID: "b-p"},
	{A: "fan", B: "pan", IPAA: "/fæn/", IPAB: "/pæn/", ContrastID: "f-p"},
	{A: "bat", B: "back", IPAA: "/bæt/", IPAB: "/bæk/", ContrastID: "final-t-k"},
	{A: "cat", B: "cap", IPAA: "/kæt/", IPAB: "/kæp/", ContrastID: "final-t-p"},
	{A: "walked", B: "walk", IPAA: "/wɔːkt/", IPAB: "/wɔːk/", ContrastID: "final-ed"},
	{A: "missed", B: "miss", IPAA: "/mɪst/", IPAB: "/mɪs/", ContrastID: "final-ed"},
}

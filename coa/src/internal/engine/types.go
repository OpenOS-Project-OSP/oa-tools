package engine

// BootloaderRoot definisce dove vengono estratti i bootloader.
const BootloaderRoot = "/tmp/coa/bootloaders"

// Struttura che mappa esattamente i campi che il C cercherà nell'array "users"
type UserDef struct {
	Login    string   `json:"login"`
	Password string   `json:"password"`
	Home     string   `json:"home"`
	Shell    string   `json:"shell"`
	Gecos    string   `json:"gecos"`
	Uid      int      `json:"uid"`
	Gid      int      `json:"gid"`
	Groups   []string `json:"groups"` // <-- L'array dei gruppi!
}

// Action rappresenta un singolo blocco "command" nell'array "plan"
type Action struct {
	Command         string   `json:"command"`
	Info            string   `json:"info,omitempty"`
	VolID           string   `json:"volid,omitempty"`
	OutputISO       string   `json:"output_iso,omitempty"`
	CryptedPassword string   `json:"crypted_password,omitempty"`
	RunCommand      string   `json:"run_command,omitempty"`
	Chroot          bool     `json:"chroot,omitempty"`
	ExcludeList     string   `json:"exclude_list,omitempty"`
	BootParams      string   `json:"boot_params,omitempty"`
	Args            []string `json:"args,omitempty"`
	// campi per l'installazione
	Device     string              `json:"device,omitempty"`
	Label      string              `json:"label,omitempty"`
	Partitions []map[string]string `json:"partitions,omitempty"`
	Actions    []map[string]string `json:"actions,omitempty"` // oa_format cerca "actions"

	// ---> I CAMPI CHE MANCAVANO QUI <---
	// Permettono a questa specifica azione di trasportare la lista utenti e la modalità
	Mode  string    `json:"mode,omitempty"`
	Users []UserDef `json:"users,omitempty"`
}

// FlightPlan è l'oggetto JSON principale inviato al motore oa
type FlightPlan struct {
	PathLiveFs      string   `json:"pathLiveFs"`
	Mode            string   `json:"mode"`
	Family          string   `json:"family"`
	InitrdCmd       string   `json:"initrd_cmd"`
	BootloadersPath string   `json:"bootloaders_path"`
	Plan            []Action `json:"plan"`
}

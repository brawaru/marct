package launcher

const (
	// Path where asset indexes reside.
	//
	// It is followed by /{indexName}
	//
	// Where:
	//
	// - indexName is name of the index file consisting of its ID and .json extension.
	assetIndexesPath = "assets/indexes"
	// Path where logging configurations reside.
	//
	// It is followed by /{logConfigName}
	//
	// Where:
	//
	// - logConfigName is name of the logging configuration file.
	logConfigsPath = "assets/log_configs"
	assetsPath     = "assets"
	// Path where asset objects reside.
	//
	// It is followed by /{assetInitials}/{hash}
	//
	// Where:
	//
	// - assetInitials are two first characters of the hash
	//
	// - hash asset hash sum.
	assetsObjectsPath = assetsPath + "/objects"
	// Path where assets are virtualised.
	//
	// It is followed by /{indexID}/{assetPath}
	//
	// Where:
	//
	// - indexID is identifier of the asset index file
	//
	// - assetPath is path of the asset.
	assetsVirtualPath = assetsPath + "/virtual"
	// Path where JREs should be installed.
	//
	// It is followed by /{classifier}/{selector}
	//
	// Where:
	//
	// - classifier is a classifier of the JRE (e.g. java-runtime-alpha)
	//
	// - selector is an appropriate system/arch selector (e.g. windows-x64)
	//
	// Resulting path is a folder with multiple files:
	//
	// - .version - a text file containing that JRE version.
	//
	// - {classifier}.sha1 - a special text file containing hash sums. // TODO: write a parser for that file
	//
	// - {classifier} - a folder containing the files of that JRE.
	runtimesPath = "runtime"
)

var LauncherIcons = []string{
	"Bedrock",
	"Bookshelf",
	"Brick",
	"Cake",
	"Carved_Pumpkin",
	"Chest",
	"Clay",
	"Coal_Block",
	"Coal_Ore",
	"Cobblestone",
	"Crafting_Table",
	"Creeper_Head",
	"Diamond_Block",
	"Diamond_Ore",
	"Dirt",
	"Dirt_Podzol",
	"Dirt_Snow",
	"Emerald_Block",
	"Emerald_Ore",
	"Enchanting_Table",
	"End_Stone",
	"Farmland",
	"Furnace",
	"Furnace_On",
	"Glass",
	"Glazed_Terracotta_Light_Blue",
	"Glazed_Terracotta_Orange",
	"Glazed_Terracotta_White",
	"Glowstone",
	"Gold_Block",
	"Gold_Ore",
	"Grass",
	"Gravel",
	"Hardened_Clay",
	"Ice_Packed",
	"Iron_Block",
	"Iron_Ore",
	"Lapis_Ore",
	"Leaves_Birch",
	"Leaves_Jungle",
	"Leaves_Oak",
	"Leaves_Spruce",
	"Lectern_Book",
	"Log_Acacia",
	"Log_Birch",
	"Log_DarkOak",
	"Log_Jungle",
	"Log_Oak",
	"Log_Spruce",
	"Mycelium",
	"Nether_Brick",
	"Netherrack",
	"Obsidian",
	"Planks_Acacia",
	"Planks_Birch",
	"Planks_DarkOak",
	"Planks_Jungle",
	"Planks_Oak",
	"Planks_Spruce",
	"Quartz_Ore",
	"Red_Sand",
	"Red_Sandstone",
	"Redstone_Block",
	"Redstone_Ore",
	"Sand",
	"Sandstone",
	"Skeleton_Skull",
	"Snow",
	"Soul_Sand",
	"Stone",
	"Stone_Andesite",
	"Stone_Diorite",
	"Stone_Granite",
	"TNT",
	"Water",
	"Wool",
}

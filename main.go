// Copyright 2023 Daniel Erat.
// All rights reserved.

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strings"
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage: %v [flag]...\n"+
			"Queries the Bandcamp Discover API and prints album URLs.\n\n", os.Args[0])
		flag.PrintDefaults()
	}
	genre := flag.String("genre", "all", "Genre or genre/subgenre to query")
	listGenres := flag.Bool("list-genres", false, "Print all genres to stdout")
	ranking := flag.String("ranking", "top", "Ranking to display (top, new, rec)")
	format := flag.String("format", "all", "Format to display (all, digital, vinyl, cd, cassette)")
	flag.Parse()

	if *listGenres {
		printGenres(os.Stdout)
		os.Exit(0)
	}

	var subgenre string
	if parts := strings.Split(*genre, "/"); len(parts) == 2 {
		*genre, subgenre = parts[0], parts[1]
	} else if len(parts) != 1 {
		fmt.Fprintln(os.Stderr, "-genre value should contain genre or genre/subgenre")
		os.Exit(2)
	}
	// TODO: Print a warning if the genre or subgenre are unknown?
	// The API looks like it just ignores invalid parameters.

	urls, err := getURLs(*genre, subgenre, *ranking, *format)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed getting URLs:", err)
		os.Exit(1)
	}

	for _, u := range urls {
		fmt.Println(u)
	}
}

func getURLs(genre, subgenre, ranking, format string) ([]string, error) {
	u := "https://bandcamp.com/api/discover/3/get_web?" +
		"g=" + genre + "&s=" + ranking + "&f=" + format + "&p=0&gn=0&w=0"
	if subgenre != "" {
		u += "&t=" + subgenre
	}
	resp, err := http.Get(u)
	if err != nil {
		return nil, fmt.Errorf("%v: %v", u, err)
	}
	defer resp.Body.Close()

	var data struct {
		Items []struct {
			PrimaryText   string `json:"primary_text"`   // album
			SecondaryText string `json:"secondary_text"` // artist
			URLHints      struct {
				Subdomain string `json:"subdomain"` // <subdomain>.bandcamp.com
				Slug      string `json:"slug"`      // /album/<slug>
				ItemType  string `json:"item_type"` // "a" for album
			} `json:"url_hints"`
		} `json:"items"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	var urls []string
	for _, item := range data.Items {
		// TODO: Do tracks use "t"?
		uh := &item.URLHints
		if uh.ItemType != "a" {
			continue
		}
		// TODO: Probably need to handle custom domains too.
		urls = append(urls, fmt.Sprintf("https://%v.bandcamp.com/album/%v", uh.Subdomain, uh.Slug))
	}
	return urls, nil
}

// printGenres prints genres (followed by indented subgenres) to w.
func printGenres(w io.Writer) {
	genres := make([]string, 0, len(allGenres))
	for g := range allGenres {
		genres = append(genres, g)
	}
	sort.Strings(genres)

	for _, g := range genres {
		fmt.Fprintln(w, g)
		for _, s := range allGenres[g] {
			fmt.Fprintln(w, "  "+s)
		}
	}
}

// allGenres maps from genres to subgenres.
// The map contents were generated by running the following in the
// JS console after loading https://bandcamp.com/#discover:
//
// el = document.getElementById('pagedata');
// data = JSON.parse(el.getAttribute('data-blob')).discover_2015.options.t;
// Object.entries(data).map(([g, subs]) => {
// 	 const s = subs.map(s => `"${s.value}",`).join("\n");
// 	 return `"${g}": []string{\n${s}\n},`
// }).join("\n");
var allGenres = map[string][]string{
	"acoustic": []string{
		"all-acoustic",
		"folk",
		"singer-songwriter",
		"rock",
		"pop",
		"guitar",
		"americana",
		"electro-acoustic",
		"instrumental",
		"piano",
		"bluegrass",
		"roots",
	},
	"alternative": []string{
		"all-alternative",
		"indie-rock",
		"industrial",
		"shoegaze",
		"grunge",
		"goth",
		"dream-pop",
		"emo",
		"math-rock",
		"britpop",
		"jangle-pop",
	},
	"ambient": []string{
		"all-ambient",
		"chill-out",
		"drone",
		"dark-ambient",
		"electronic",
		"soundscapes",
		"field-recordings",
		"atmospheric",
		"meditation",
		"noise",
		"new-age",
		"idm",
		"industrial",
	},
	"blues": []string{
		"all-blues",
		"rhythm-blues",
		"blues-rock",
		"country-blues",
		"boogie-woogie",
		"delta-blues",
		"americana",
		"electric-blues",
		"gospel",
		"bluegrass",
	},
	"classical": []string{
		"all-classical",
		"orchestral",
		"neo-classical",
		"chamber-music",
		"classical-piano",
		"contemporary-classical",
		"baroque",
		"opera",
		"choral",
		"modern-classical",
		"avant-garde",
	},
	"comedy": []string{
		"all-comedy",
		"improv",
		"stand-up",
	},
	"country": []string{
		"all-country",
		"bluegrass",
		"country-rock",
		"americana",
		"country-folk",
		"alt-country",
		"country-blues",
		"western",
		"singer-songwriter",
		"outlaw",
		"honky-tonk",
		"roots",
		"hillbilly",
	},
	"devotional": []string{
		"all-devotional",
		"christian",
		"gospel",
		"meditation",
		"spiritual",
		"worship",
		"inspirational",
	},
	"electronic": []string{
		"all-electronic",
		"house",
		"electronica",
		"downtempo",
		"techno",
		"electro",
		"dubstep",
		"beats",
		"dance",
		"idm",
		"drum-bass",
		"breaks",
		"trance",
		"glitch",
		"chiptune",
		"chillwave",
		"dub",
		"edm",
		"instrumental",
		"witch-house",
		"garage",
		"juke",
		"footwork",
		"vaporwave",
		"synthwave",
	},
	"experimental": []string{
		"all-experimental",
		"noise",
		"drone",
		"avant-garde",
		"experimental-rock",
		"improvisation",
		"sound-art",
		"musique-concrete",
	},
	"folk": []string{
		"all-folk",
		"singer-songwriter",
		"folk-rock",
		"indie-folk",
		"pop-folk",
		"traditional",
		"experimental-folk",
		"roots",
	},
	"funk": []string{
		"all-funk",
		"funk-jam",
		"deep-funk",
		"funk-rock",
		"jazz-funk",
		"boogie",
		"g-funk",
		"rare-groove",
		"electro",
		"go-go",
	},
	"hip-hop-rap": []string{
		"all-hip-hop-rap",
		"rap",
		"underground-hip-hop",
		"instrumental-hip-hop",
		"trap",
		"conscious-hip-hop",
		"boom-bap",
		"beat-tape",
		"hardcore",
		"grime",
	},
	"jazz": []string{
		"all-jazz",
		"fusion",
		"big-band",
		"nu-jazz",
		"modern-jazz",
		"swing",
		"free-jazz",
		"soul-jazz",
		"latin-jazz",
		"vocal-jazz",
		"bebop",
		"spiritual-jazz",
	},
	"kids": []string{
		"all-kids",
		"family-music",
		"educational",
		"music-therapy",
		"lullaby",
		"baby",
	},
	"latin": []string{
		"all-latin",
		"brazilian",
		"cumbia",
		"tango",
		"latin-rock",
		"flamenco",
		"salsa",
		"reggaeton",
		"merengue",
		"bolero",
		"méxico-d.f.",
		"bachata",
	},
	"metal": []string{
		"all-metal",
		"hardcore",
		"black-metal",
		"death-metal",
		"thrash-metal",
		"grindcore",
		"doom",
		"post-hardcore",
		"progressive-metal",
		"metalcore",
		"sludge-metal",
		"heavy-metal",
		"deathcore",
		"noise",
	},
	"pop": []string{
		"all-pop",
		"indie-pop",
		"synth-pop",
		"power-pop",
		"new-wave",
		"dream-pop",
		"noise-pop",
		"experimental-pop",
		"electro-pop",
		"adult-contemporary",
		"jangle-pop",
		"j-pop",
	},
	"punk": []string{
		"all-punk",
		"hardcore-punk",
		"garage",
		"pop-punk",
		"punk-rock",
		"post-punk",
		"post-hardcore",
		"thrash",
		"crust-punk",
		"folk-punk",
		"emo",
		"ska",
		"no-wave",
	},
	"r-b-soul": []string{
		"all-r-b-soul",
		"soul",
		"r-b",
		"neo-soul",
		"gospel",
		"contemporary-r-b",
		"motown",
		"urban",
	},
	"reggae": []string{
		"all-reggae",
		"dub",
		"ska",
		"roots",
		"dancehall",
		"rocksteady",
		"ragga",
		"lovers-rock",
	},
	"rock": []string{
		"all-rock",
		"indie",
		"prog-rock",
		"post-rock",
		"rock-roll",
		"psychedelic-rock",
		"hard-rock",
		"garage-rock",
		"surf-rock",
		"instrumental",
		"math-rock",
		"rockabilly",
	},
	"soundtrack": []string{
		"all-soundtrack",
		"film-music",
		"video-game-music",
	},
	"spoken-word": []string{
		"all-spoken-word",
		"poetry",
		"inspirational",
		"storytelling",
		"self-help",
	},
	"world": []string{
		"all-world",
		"latin",
		"roots",
		"african",
		"tropical",
		"tribal",
		"brazilian",
		"celtic",
		"world-fusion",
		"cumbia",
		"gypsy",
		"new-age",
		"balkan",
		"reggaeton",
	},
}

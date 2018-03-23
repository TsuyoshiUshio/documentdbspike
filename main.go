package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"sort"
	"time"

	documentdb "github.com/TsuyoshiUshio/documentdb-go"
)

type Service struct {
	Name  string
	Value int
}
type Team struct {
	documentdb.Document
	Name     string
	Services *[]Service
}

func (c *Team) Update() {
	ch := make(chan Service)

	for _, v := range *c.Services {
		go func(service Service) {
			service.Update()
			ch <- service
		}(v)
	}
	var services []Service
	for i := 0; i < len(*c.Services); i++ {
		result := <-ch
		services = append(services, result)
	}
	c.Services = &services
}

func (s *Service) Update() {
	s.Value = s.Value + 1
	time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond)
}

func GoRoutineWithoutChannelWithSort(teams *[]Team) *[]Team {
	ch := make(chan Team)
	for _, v := range *teams {
		go func(team Team) {
			team.Update()
			ch <- team
		}(v)
	}
	var newTeams []Team
	for i := 0; i < len(*teams); i++ {
		result := <-ch
		newTeams = append(newTeams, result)
	}

	sort.Slice(newTeams, func(i, j int) bool {
		return (newTeams[i].Name < newTeams[j].Name)
	})

	return &newTeams
}

// DB interface
type DB interface {
	Get(id string) *Team
	GetAll() []*Team
	Add(u *Team) *Team
	Update(u *Team) *Team
	Remove(id string) error
}

// UsersDB implement DB interface
type TeamDB struct {
	Database   string
	Collection string
	db         *documentdb.Database
	coll       *documentdb.Collection
	client     *documentdb.DocumentDB
}

// DocumentDB config
type Config struct {
	Url       string `json:"url"`
	MasterKey string `json:"masterKey"`
}

// Return new UserDB
// Test if database and collection exist. if not, create them.
func NewDB(db, coll string, config *Config) (teamdb TeamDB) {
	teamdb.Database = db
	teamdb.Collection = coll
	teamdb.client = documentdb.New(config.Url, documentdb.Config{config.MasterKey})
	// Find or create `test` db and `users` collection
	if err := teamdb.findOrDatabase(db); err != nil {
		panic(err)
	}
	if err := teamdb.findOrCreateCollection(coll); err != nil {
		panic(err)
	}
	return
}

// Find or create collection by id
func (u *TeamDB) findOrCreateCollection(name string) (err error) {
	if colls, err := u.client.QueryCollections(u.db.Self, fmt.Sprintf("SELECT * FROM ROOT r WHERE r.id='%s'", name)); err != nil {
		return err
	} else if len(colls) == 0 {
		if coll, err := u.client.CreateCollection(u.db.Self, fmt.Sprintf(`{ "id": "%s" }`, name)); err != nil {
			return err
		} else {
			u.coll = coll
		}
	} else {
		u.coll = &colls[0]
	}
	return
}

// Find or create database by id
func (u *TeamDB) findOrDatabase(name string) (err error) {
	if dbs, err := u.client.QueryDatabases(fmt.Sprintf("SELECT * FROM ROOT r WHERE r.id='%s'", name)); err != nil {
		return err
	} else if len(dbs) == 0 {
		if db, err := u.client.CreateDatabase(fmt.Sprintf(`{ "id": "%s" }`, name)); err != nil {
			return err
		} else {
			u.db = db
		}
	} else {
		u.db = &dbs[0]
	}
	return
}

// Get all users
func (u *TeamDB) GetAll() (teams []Team, err error) {
	err = u.client.ReadDocuments(u.coll.Self, &teams)
	return
}

// Create user
func (u *TeamDB) Add(team *Team) (err error) {
	jsonbytes, err := json.Marshal(*team)
	if err != nil {
		panic(err)
	}
	fmt.Println("--------")
	fmt.Println(string(jsonbytes))
	return u.client.UpsertDocument(u.coll.Self, team)
}

// func upsertWithMongoAPI() {
// dialInfo := &mgo.DialInfo{
// 	Addrs:    []string{"host"},
// 	Timeout:  60 * time.Second,
// 	Database: "database",
// 	Username: "username",
// 	Password: "password",
// 	DialServer: func(addr *mgo.ServerAddr) (net.Conn, error) {
// 		return tls.Dial("tcp", addr.String(), &tls.Config{})
// 	},
// }
//}

func main() {
	data, err := ioutil.ReadFile("config.json")
	if err != nil {
		panic(err)
	}
	config := &Config{}
	if err = json.Unmarshal(data, config); err != nil {
		panic(err)
	}

	teamdb := NewDB("sadb", "col", config)

	// Create an sample data by a struct
	teams := Setup()
	// Marshal into Json (Don't need it)
	// Upsert all of them
	for i, v := range *teams {
		err := teamdb.Add(&v)
		if err != nil {
			fmt.Printf("upsert error! %d : %v", i, err)
		}
	}
	// Read all of them

	newTeams, err := teamdb.GetAll()
	if err != nil {
		panic(err)
	}

	// Update the model
	updatedTeams := GoRoutineWithoutChannelWithSort(&newTeams)

	// Upsert the database
	for i, v := range *updatedTeams {
		err := teamdb.Add(&v)
		errString := err.Error()
		if err != nil {
			// If you use upsert, the error happens with error sting ', '
			// which means status code validation error. Upsert can't predict which status code 200 or 201 in advance.
			if ", " != errString {
				fmt.Printf("upsert error! %d : '%s'", i, err.Error())
			}
		}
	}
	// Getting a benchmark
	fmt.Println("Done...")
	//

}

func Setup() *[]Team {
	//	var teams []Team = make([]Team, 20)
	var teams []Team
	for i := 0; i < 20; i++ {
		team := Team{
			Name: "Team " + fmt.Sprint(i),
		}
		// services := make([]Service, 5)
		var services []Service
		for j := 0; j < 5; j++ {
			service := Service{
				Name: "Servicer " + fmt.Sprint(j),
			}
			services = append(services, service)
		}
		team.Services = &services
		teams = append(teams, team)
	}
	return &teams
}

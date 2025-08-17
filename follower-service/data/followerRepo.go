package data

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

type FollowerRepo struct {
	driver neo4j.DriverWithContext
	logger *log.Logger
}

func New(logger *log.Logger) (*FollowerRepo, error) {
	uri := os.Getenv("NEO4J_URI")
	username := os.Getenv("NEO4J_USERNAME")
	password := os.Getenv("NEO4J_PASSWORD")
	auth := neo4j.BasicAuth(username, password, "")

	driver, err := neo4j.NewDriverWithContext(uri, auth)
	if err != nil {
		logger.Panic(err)
		return nil, err
	}

	return &FollowerRepo{
		driver: driver,
		logger: logger,
	}, nil
}

func (r *FollowerRepo) CheckConnection() {
	ctx := context.Background()
	err := r.driver.VerifyConnectivity(ctx)
	if err != nil {
		r.logger.Panic(err)
		return
	}
	r.logger.Printf("Successfully connected to Neo4j database: %s", r.driver.Target().Host)
	r.CreateConstraints()
}

func (r *FollowerRepo) CreateConstraints() {
	ctx := context.Background()
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: "neo4j"})
	defer session.Close(ctx)

	// Create constraints for User nodes
	_, err := session.ExecuteWrite(ctx,
		func(transaction neo4j.ManagedTransaction) (any, error) {
			_, err := transaction.Run(ctx,
				"CREATE CONSTRAINT id IF NOT EXISTS FOR (u:User) REQUIRE u.id IS UNIQUE",
				nil)
			if err != nil {
				return nil, err
			}
			_, err = transaction.Run(ctx,
				"CREATE CONSTRAINT id IF NOT EXISTS FOR (u:User) REQUIRE u.username IS UNIQUE",
				nil)
			if err != nil {
				return nil, err
			}
			return nil, nil
		})
	if err != nil {
		r.logger.Println("Error creating constraints:", err)
		return
	}
	r.logger.Println("Constraints created successfully")
}

func (r *FollowerRepo) CloseDriverConnection(ctx context.Context) {
	fmt.Println("Closing Neo4j driver connection")
	if err := r.driver.Close(ctx); err != nil {
		r.logger.Panic(err)
	}
}

func (r *FollowerRepo) IsFollowedByMe(followerId int, followedId int) (Users, error) {
	ctx := context.Background()
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: "neo4j"})
	defer session.Close(ctx)

	// ExecuteRead for read transactions (Read and queries)
	userResults, err := session.ExecuteRead(ctx,
		func(transaction neo4j.ManagedTransaction) (any, error) {
			result, err := transaction.Run(ctx,
				`MATCH (u:User)-[:FOLLOWS]->(f:User) WHERE u.id = $followerId AND f.id = $followedId RETURN f.id AS id, f.username AS username`,
				map[string]any{"followerId": followerId, "followedId": followedId})
			if err != nil {
				return nil, err
			}

			var users Users
			for result.Next(ctx) {
				record := result.Record()
				id, _ := record.Get("id")
				username, _ := record.Get("username")
				users = append(users, &User{
					ID:       (int)(id.(int64)),
					Username: username.(string),
				})
			}
			return users, nil
		})
	if err != nil {
		r.logger.Println("Error querying search:", err)
		return nil, err
	}
	return userResults.(Users), nil
}

func (r *FollowerRepo) GetFollowed(userId int) (Users, error) {
	ctx := context.Background()
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: "neo4j"})
	defer session.Close(ctx)

	// ExecuteRead for read transactions (Read and queries)
	userResults, err := session.ExecuteRead(ctx,
		func(transaction neo4j.ManagedTransaction) (any, error) {
			result, err := transaction.Run(ctx,
				`MATCH (u:User)-[:FOLLOWS]->(f:User) WHERE u.id = $userId RETURN f.id AS id, f.username AS username`,
				map[string]any{"userId": userId})
			if err != nil {
				return nil, err
			}

			var users Users
			for result.Next(ctx) {
				record := result.Record()
				id, _ := record.Get("id")
				username, _ := record.Get("username")
				users = append(users, &User{
					ID:       (int)(id.(int64)),
					Username: username.(string),
				})
			}
			return users, nil
		})
	if err != nil {
		r.logger.Println("Error querying search:", err)
		return nil, err
	}
	return userResults.(Users), nil
}

func (r *FollowerRepo) GetFollowers(userId int) (Users, error) {
	ctx := context.Background()
	session := r.driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: "neo4j"})
	defer session.Close(ctx)

	// ExecuteRead for read transactions (Read and queries)
	userResults, err := session.ExecuteRead(ctx,
		func(transaction neo4j.ManagedTransaction) (any, error) {
			result, err := transaction.Run(ctx,
				`MATCH (u:User)-[:FOLLOWS]->(f:User) WHERE f.id = $userId RETURN u.id AS id, u.username AS username`,
				map[string]any{"userId": userId})
			if err != nil {
				return nil, err
			}

			var users Users
			for result.Next(ctx) {
				record := result.Record()
				id, _ := record.Get("id")
				username, _ := record.Get("username")
				users = append(users, &User{
					ID:       (int)(id.(int64)),
					Username: username.(string),
				})
			}
			return users, nil
		})
	if err != nil {
		r.logger.Println("Error querying search:", err)
		return nil, err
	}
	return userResults.(Users), nil
}

func (pr *FollowerRepo) Follow(follower *User, followed *User) (user *User, err error) {
	ctx := context.Background()
	session := pr.driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: "neo4j"})
	defer session.Close(ctx)

	// ExecuteWrite for write transactions (Create/Update/Delete)
	returnedValue, err := session.ExecuteWrite(ctx,
		func(transaction neo4j.ManagedTransaction) (any, error) {
			result, err := transaction.Run(ctx,
				"MERGE (a:User {id: $followerId, username: $followerUsername}) MERGE (b:User {id: $followedId, username: $followedUsername}) MERGE (a)-[r:FOLLOWS]->(b) RETURN b.id AS id, b.username AS username",
				map[string]any{"followerId": follower.ID, "followerUsername": follower.Username, "followedId": followed.ID, "followedUsername": followed.Username})
			if err != nil {
				return nil, err
			}
			var user *User
			for result.Next(ctx) {
				record := result.Record()
				id, _ := record.Get("id")
				username, _ := record.Get("username")
				user = &User{
					ID:       int(id.(int64)),
					Username: username.(string),
				}
			}

			return user, result.Err()
		})
	if err != nil {
		pr.logger.Println("Error following user:", err)
		return nil, err
	}
	pr.logger.Println(returnedValue.(*User))
	followedUser := returnedValue.(*User)
	if followedUser == nil {
		return nil, fmt.Errorf("type assertion to *User failed")
	}
	return followedUser, nil
}

func (pr *FollowerRepo) Unfollow(follower *User, unfollowed *User) (user string, err error) {
	ctx := context.Background()
	session := pr.driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: "neo4j"})
	defer session.Close(ctx)

	// ExecuteWrite for write transactions (Create/Update/Delete)
	returnedValue, err := session.ExecuteWrite(ctx,
		func(transaction neo4j.ManagedTransaction) (any, error) {
			result, err := transaction.Run(ctx,
				"MATCH (a:User) - [r:FOLLOWS]-(b:User) WHERE a.id = $followerId AND b.id = $unfollowedId DELETE r RETURN b.username AS username",
				map[string]any{"followerId": follower.ID, "unfollowedId": unfollowed.ID})
			if err != nil {
				return nil, err
			}

			if result.Next(ctx) {
				return result.Record().Values[0], nil
			}

			return nil, result.Err()
		})
	if err != nil {
		pr.logger.Println("Error unfollowing User:", err)
		return "", err
	}
	if returnedValue == nil {
		pr.logger.Println("No follow relationship found")
		return "", fmt.Errorf("no follow relationship found")
	}
	pr.logger.Println(returnedValue.(string))
	unfollowedUser := returnedValue.(string)
	if unfollowedUser == "" {
		return "", fmt.Errorf("type assertion to *User failed")
	}
	return unfollowedUser, nil
}

package db

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"social/internal/store"
)

var usernames = []string{"user", "player", "dev", "admin", "member", "guest", "pro", "master", "hero", "star"}
var passwords = []string{"123", "x", "pro", "99", "2024", "cool", "zone", "king", "lord", "ace"}
var emails = []string{"gmail.com", "yahoo.com", "hotmail.com", "outlook.com", "example.com"}
var titles = []string{
	"10 AI Tools That Will Change How You Work in 2026",
	"Why Your Smart Home Might Be Smarter Than You Think",
	"The Hidden Security Risks of Public Wi-Fi and How to Stay Safe",
	"Is Coding Still Worth Learning? An Honest Guide",
	"Morning Routines of Highly Productive People That Actually Work",
	"Why Sleep Tracking Might Be Harming Your Sleep",
	"The Minimalist Wardrobe: Less Clothes, More Confidence",
	"Digital Detox: A Weekend Without Screens – My Honest Experience",
	"The Art of Saying No Without Feeling Guilty",
	"How I Turned My Bad Habits Into Powerful Routines",
	"Why Failure Is the Best Teacher Nobody Talks About",
	"Stop Comparing: A Guide to Mindful Social Media Use",
	"Side Hustles That Actually Pay in 2026",
	"How to Ask for a Raise Without Feeling Awkward",
	"Investing for Beginners: Start With Less Than $100",
	"What Your Handwriting Says About Your Personality",
	"The Best Books That Changed My Perspective on Life",
	"Hidden Gems: Underrated Travel Destinations for 2026",
	"Why I Stopped Perfectionism and Started Creating",
	"The Psychology Behind Why We Procrastinate",
}
var contents = []string{
	"Discover the most powerful AI tools that are revolutionizing the workplace. From automation to creative assistance, these tools can boost your productivity by 10x and help you stay ahead of the competition.",
	"Smart home devices are becoming increasingly sophisticated, but are they really making our lives easier? Explore the surprising ways your connected devices might be outsmarting you.",
	"Public Wi-Fi is convenient but dangerous. Learn about the hidden security threats lurking in coffee shops, airports, and hotels, and discover practical steps to protect your personal data.",
	"Coding has never been more accessible, but is it still a valuable skill? We break down the current job market, salary trends, and alternative career paths to help you make an informed decision.",
	"Wake up earlier, work harder, be more productive. But what if the secret to success isn't about adding more to your routine, but removing the things that drain your energy?",
	"Sleep tracking apps promise to help you rest better, but research suggests they might be causing anxiety and disrupting your natural sleep patterns. Here's what the science says.",
	"Declutter your closet and transform your daily routine. A minimalist wardrobe can save you time, money, and mental energy while helping you look and feel your best.",
	"I spent 48 hours completely disconnected from all screens. No phone, no laptop, no TV. This is what happened to my mind, my productivity, and my relationships.",
	"Every time you say yes to something unimportant, you're saying no to something that matters. Learn how to set boundaries without damaging your relationships or career.",
	"For years I struggled with bad habits. Then I discovered a simple framework that completely transformed my daily routine. Here's exactly what I did and how you can too.",
	"Failure has a bad reputation. We celebrate success stories but rarely talk about the countless failures that paved the way. It's time to reframe how we think about mistakes.",
	"Social media was supposed to connect us, but somehow we feel more isolated than ever. Learn how to take control of your feed and use platforms intentionally.",
	"The gig economy is booming, but not all side hustles are created equal. After testing dozens of options, here are the ones that actually generate reliable income.",
	"Asking for more money feels uncomfortable, but it shouldn't. With the right approach and timing, you can negotiate your salary confidently and professionally.",
	"You don't need thousands of dollars to start investing. Even $50 a month in the right index funds can grow into substantial wealth over time. Here's how to begin.",
	"Your handwriting reveals more about your personality than you might think. Graphologists have studied the link between writing style and character traits for centuries.",
	"Some books completely reshape how you see the world. These transformative reads combine practical wisdom with compelling storytelling that stays with you forever.",
	"Skip the tourist traps and discover breathtaking destinations that most travelers overlook. These hidden gems offer authentic experiences without the crowds.",
	"Perfectionism held me back for years. I was so afraid of making mistakes that I never shipped any projects. Here's how I learned to embrace imperfection.",
	"We all procrastinate, but understanding why can help you overcome it. Explore the psychological triggers behind delay and learn science-backed strategies to beat them.",
}
var tags = []string{
	"technology",
	"ai",
	"productivity",
	"lifestyle",
	"career",
	"finance",
	"self-improvement",
	"health",
	"sleep",
	"minimalism",
	"digital-detox",
	"side-hustle",
	"investing",
	"books",
	"travel",
	"psychology",
	"habits",
	"mindfulness",
	"work-life-balance",
	"creativity",
}
var commentsContent = []string{
	"Great article! I never thought about it that way. Shared with my team.",
	"This is exactly what I needed to read today. Thank you for sharing!",
	"I completely agree with this. Been following your blog for months.",
	"Very informative! Could you write a follow-up post on this topic?",
	"Interesting perspective. I had a different experience but I see your point.",
	"This helped me a lot! I implemented some of these tips and they work.",
	"Love the writing style. Very easy to read and understand.",
	"I've been looking for content like this. Bookmarked for later!",
	"Would love to see more posts on this subject. Please keep writing!",
	"This changed my mindset completely. Thank you for the valuable insights.",
	"Shared on LinkedIn. I think more people should read this.",
	"Finally someone said it! I've been thinking the same thing for years.",
	"Quick question: do you have any resources you can recommend for this?",
	"Amazing! This is so helpful for beginners like me. Keep it up!",
	"I tried this approach and it worked wonders for my productivity.",
	"Great tips! I especially liked the part about digital detox.",
	"This is such an underrated topic. Glad you covered it.",
	"Wow, I learned something new today. Looking forward to more content!",
	"The examples you provided made it so much easier to understand.",
	"Thanks for breaking this down. Most articles skip the important details.",
	"Absolutely brilliant! You've gained a new follower here.",
	"This should be required reading for everyone in my industry.",
	"I appreciate how you kept it practical and not just theory.",
	"Can't wait to try these suggestions this weekend. Will report back!",
	"You always deliver quality content. This is why I keep coming back!",
	"Exactly the motivation I needed today. Thank you so much!",
	"I wish I found this article sooner. This is gold!",
	"The structure of this post is perfect. Very well organized.",
	"Hit the nail on the head with this one. Well done!",
	"Doing this changed my life. Thanks for the inspiration!",
}

func Seed(store store.Storage, db *sql.DB) {
	ctx := context.Background()
	users := generateUsers(100)
	tx, _ := db.BeginTx(ctx, nil)

	for _, user := range users {
		if err := store.Users.Create(ctx, tx, user); err != nil {
			_ = tx.Rollback()
			log.Println("Failed to create user:", err)
		}
	}
	_ = tx.Commit()
	posts := generatePosts(300, users)
	for _, post := range posts {
		if err := store.Posts.Create(ctx, post); err != nil {
			log.Println("Failed to create post:", err)
		}
	}
	commentsUser := generateComments(100, posts, users)
	for _, comments := range commentsUser {
		if err := store.Comments.Create(ctx, comments); err != nil {
			log.Println("Failed to create comment:", err)
		}
	}

	for range 100 {
		user := users[rand.Intn(len(users))]
		follower := users[rand.Intn(len(users))]

		if user.ID == follower.ID {
			continue
		}
		if err := store.Followers.Follow(ctx, follower.ID, user.ID); err != nil {
			log.Println("Failed to follow follower:", err)
		}
	}
	log.Println("Seeding is end")
}

func generateUsers(num int) []*store.User {
	users := make([]*store.User, num)
	for i := range num {
		users[i] = &store.User{
			Username: usernames[i%len(usernames)] + fmt.Sprintf("%d", i),
			Email:    usernames[i%len(usernames)] + fmt.Sprintf("%d", i) + "@" + emails[rand.Intn(len(emails))],
		}
		users[i].Password.Set("123")
	}

	return users
}
func generatePosts(num int, users []*store.User) []*store.Post {
	posts := make([]*store.Post, num)
	for i := range num {
		user := users[rand.Intn(len(users))]
		posts[i] = &store.Post{
			UserID:  user.ID,
			Title:   titles[rand.Intn(len(titles))],
			Content: contents[rand.Intn(len(contents))],
			Tags:    []string{tags[rand.Intn(len(tags))], tags[rand.Intn(len(tags))]},
		}
	}
	return posts
}
func generateComments(num int, posts []*store.Post, users []*store.User) []*store.Comment {
	comments := make([]*store.Comment, num)
	for i := range num {
		user := users[rand.Intn(len(users))]
		post := posts[rand.Intn(len(posts))]
		comments[i] = &store.Comment{
			PostID:  post.ID,
			UserID:  user.ID,
			Content: commentsContent[rand.Intn(len(commentsContent))],
			User:    *user,
		}
	}
	return comments
}

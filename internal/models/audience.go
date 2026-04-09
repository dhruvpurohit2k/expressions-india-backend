package models

import "gorm.io/gorm"

type Audience struct {
	ID          uint   `gorm:"primaryKey" json:"-"`
	Name        string `gorm:"uniqueIndex" json:"name"`
	Description string `gorm:"type:text" json:"introduction"`
}

func SeedAudience(db *gorm.DB) {
	options := []Audience{
		{Name: "all", Description: ""},
		{Name: "student", Description: "Expressions India is a safe, supportive space for students navigating the pressures of academics, identity, and growing up. Here you'll find curated events, articles, and conversations designed to help you understand your emotions, build resilience, and thrive — both inside and outside the classroom."},
		{Name: "teacher", Description: "Teachers are the backbone of student wellbeing. This space is designed to equip educators with the tools, insights, and community they need to recognise signs of distress, foster safe classrooms, and model the importance of mental health — for their students and for themselves."},
		{Name: "head_of_department", Description: "As a Head of Department, you shape the culture of your school. This section offers research-backed resources, policy insights, and practical guidance to help you build environments where students and staff feel seen, supported, and empowered to seek help when they need it."},
		{Name: "parent", Description: "Parenting in today's world comes with new challenges — from screen time to exam pressure to conversations about mental health. This space gives parents evidence-based guidance, expert perspectives, and community support to help raise emotionally healthy, resilient children."},
		{Name: "counselor", Description: "School counselors are on the frontlines of student mental health. This curated section brings together clinical insights, intervention frameworks, and peer conversations to support your practice — so you can continue showing up fully for the young people who need you most."},
		{Name: "mental_health_professional", Description: "Mental health professionals working with young people face a unique set of challenges and responsibilities. This space is built for practitioners who want to stay informed on adolescent psychology, evidence-based approaches, and the evolving landscape of school-based mental health support in India."},
	}
	for _, opt := range options {
		var existing Audience
		if err := db.Where("name = ?", opt.Name).First(&existing).Error; err != nil {
			db.Create(&opt)
		} else if existing.Description == "" && opt.Description != "" {
			db.Model(&existing).Update("description", opt.Description)
		}
	}
}

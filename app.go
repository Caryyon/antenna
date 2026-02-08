package main

import (
	"context"

	"github.com/Caryyon/antenna/internal/api"
)

// App struct
type App struct {
	ctx    context.Context
	client *api.Client
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{
		client: api.NewClient(""),
	}
}

// startup is called when the app starts
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// Session is re-exported for Wails bindings
type Session = api.Session

// DashboardData is re-exported for Wails bindings
type DashboardData = api.DashboardData

// HourlyBucket is re-exported for Wails bindings
type HourlyBucket = api.HourlyBucket

// GetDashboard returns the dashboard data
func (a *App) GetDashboard() DashboardData {
	return a.client.GetDashboard()
}

// GetHourlyActivity returns message counts and costs bucketed by hour for the last 24h
func (a *App) GetHourlyActivity() []HourlyBucket {
	return a.client.GetHourlyActivity()
}

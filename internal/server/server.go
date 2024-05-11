package server

import (
	"context"

	"github.com/maxmwang/goscrape/internal/proto"
	"github.com/maxmwang/goscrape/internal/scrape"
	"github.com/maxmwang/goscrape/internal/sqlc"
)

type ScraperServer struct {
	proto.UnimplementedScraperServer

	q *sqlc.Queries
}

func New(q *sqlc.Queries) *ScraperServer {
	return &ScraperServer{
		q: q,
	}
}

func (s *ScraperServer) Try(ctx context.Context, req *proto.TryRequest) (*proto.TryReply, error) {
	var exists bool
	exist, err := s.q.GetCompany(ctx, req.Company)
	if err == nil && len(exist) > 0 {
		exists = true
	}

	site, jobs, err := scrape.ScrapeAll(req.Company)
	if err != nil {
		return &proto.TryReply{}, nil
	}

	target := 0
	for _, j := range jobs {
		if j.IsTarget() {
			target++
		}
	}

	return &proto.TryReply{
		Site:   site,
		Count:  uint32(len(jobs)),
		Target: uint32(target),
		Exists: exists,
	}, nil
}

func (s *ScraperServer) TryThenAdd(ctx context.Context, req *proto.TryRequest) (*proto.TryThenAddReply, error) {
	try, err := s.Try(ctx, req)
	if err != nil {
		return nil, err
	}

	var added bool
	if try.Site != "" && !try.Exists && try.Count > 0 {
		err := s.q.AddCompany(ctx, sqlc.AddCompanyParams{
			Name: req.Company,
			Site: try.Site,
		})
		if err != nil {
			return nil, err
		}
		added = true
	}

	return &proto.TryThenAddReply{
		Site:   try.Site,
		Count:  try.Count,
		Target: try.Target,
		Exists: try.Exists,
		Added:  added,
	}, nil
}

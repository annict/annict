# frozen_string_literal: true

["", "en.", "ja."].each do |subdomain|
  SitemapGenerator::Sitemap.default_host = "https://#{subdomain}annict.com"
  SitemapGenerator::Sitemap.sitemaps_path = "sitemaps/#{subdomain.delete('.')}"

  adapter_options = {
    fog_provider: "AWS",
    aws_access_key_id: ENV.fetch("S3_ACCESS_KEY_ID"),
    aws_secret_access_key: ENV.fetch("S3_SECRET_ACCESS_KEY"),
    fog_directory: ENV.fetch("S3_BUCKET_NAME"),
    fog_region: ENV.fetch("S3_REGION")
  }
  SitemapGenerator::Sitemap.adapter = SitemapGenerator::S3Adapter.new(adapter_options)

  SitemapGenerator::Sitemap.create do
    add about_path, priority: 0.7
    add popular_works_path, priority: 0.9
    add newest_works_path, priority: 0.9
    add privacy_path
    add search_path, priority: 0.8
    add terms_path

    User.find_each do |u|
      add user_path(u.username), priority: 0.9, lastmod: u.updated_at
      add favorite_characters_path(u.username), priority: 0.8, lastmod: u.updated_at
      add favorite_people_path(u.username), priority: 0.8, lastmod: u.updated_at
      add favorite_organizations_path(u.username), priority: 0.8, lastmod: u.updated_at

      Status.kind.values.each do |k|
        add user_works_path(u.username, k), priority: 0.8
      end

      u.records.find_each do |r|
        add record_path(u.username, r), priority: 0.8
      end
    end

    Season.list do |s|
      add season_works_path(s.slug), priority: 0.9
    end

    Person.published.find_each do |p|
      if p.casts.published.present? || p.staffs.published.present?
        add person_path(p), priority: 0.8, lastmod: p.updated_at
      end
    end

    Organization.published.find_each do |o|
      if o.staffs.published.present?
        add organization_path(o), priority: 0.8, lastmod: o.updated_at
      end
    end

    Work.published.find_each do |w|
      add work_path(w.id), priority: 1.0, lastmod: w.updated_at

      w.episodes.published.find_each do |e|
        add work_episode_path(w.id, e.id), priority: 1.0 if e.records.present?
      end
    end

    Character.published.find_each do |c|
      add character_path(c.id), priority: 0.9, lastmod: c.updated_at
      add character_fans_path(c.id), priority: 0.7, lastmod: c.updated_at
    end

    Person.published.find_each do |p|
      add person_path(p.id), priority: 0.9, lastmod: p.updated_at
      add person_fans_path(p.id), priority: 0.7, lastmod: p.updated_at
    end

    Organization.published.find_each do |o|
      add organization_path(o.id), priority: 0.9, lastmod: o.updated_at
      add organization_fans_path(o.id), priority: 0.7, lastmod: o.updated_at
    end
  end
end

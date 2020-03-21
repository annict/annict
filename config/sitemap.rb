# frozen_string_literal: true

[
  { host: ENV.fetch("ANNICT_HOST"), namespace: "en" },
  { host: ENV.fetch("ANNICT_JP_HOST"), namespace: "ja" }
].each do |data|
  SitemapGenerator::Sitemap.default_host = "https://#{data[:host]}"
  SitemapGenerator::Sitemap.sitemaps_path = "sitemaps/#{data[:namespace]}"

  adapter_options = {
    aws_access_key_id: ENV.fetch("S3_ACCESS_KEY_ID"),
    aws_secret_access_key: ENV.fetch("S3_SECRET_ACCESS_KEY"),
    aws_region: ENV.fetch("S3_REGION")
  }
  SitemapGenerator::Sitemap.adapter = SitemapGenerator::AwsSdkAdapter.new(ENV.fetch("S3_BUCKET_NAME"), adapter_options)

  SitemapGenerator::Sitemap.create do
    add about_path, priority: 0.7
    add popular_works_path, priority: 0.9
    add newest_works_path, priority: 0.9
    add privacy_path
    add search_path, priority: 0.8
    add terms_path

    User.only_kept.find_each do |u|
      add user_path(u.username), priority: 0.9, lastmod: u.updated_at

      Status.kind.values.each do |k|
        add library_path(u.username, k), priority: 0.8
      end

      u.episode_records.only_kept.with_body.find_each do |er|
        add record_path(u.username, er.record), priority: 0.8
      end

      u.work_records.only_kept.with_body.find_each do |wr|
        add record_path(u.username, wr.record), priority: 0.8
      end
    end

    Season.list do |s|
      add season_works_path(s.slug), priority: 0.9
    end

    Character.only_kept.find_each do |c|
      if c.casts.only_kept.present?
        add character_path(c.id), priority: 0.9, lastmod: c.updated_at
      end
    end

    Person.only_kept.find_each do |p|
      if p.casts.only_kept.present? || p.staffs.only_kept.present?
        add person_path(p), priority: 0.8, lastmod: p.updated_at
      end
    end

    Organization.only_kept.find_each do |o|
      if o.staffs.only_kept.present?
        add organization_path(o), priority: 0.8, lastmod: o.updated_at
      end
    end

    Work.only_kept.find_each do |w|
      add work_path(w.id), priority: 1.0, lastmod: w.updated_at

      w.episodes.only_kept.find_each do |e|
        add work_episode_path(w.id, e.id), priority: 1.0 if e.episode_records.only_kept.with_body.present?
      end
    end
  end
end

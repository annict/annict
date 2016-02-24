S3_REGION = ENV.fetch("S3_REGION")
S3_BUCKET_NAME = ENV.fetch("S3_BUCKET_NAME")
HOST = "https://s3-#{S3_REGION}.amazonaws.com/#{S3_BUCKET_NAME}/"

SitemapGenerator::Sitemap.default_host = "https://annict.com"
SitemapGenerator::Sitemap.sitemaps_host = HOST
SitemapGenerator::Sitemap.sitemaps_path = "sitemaps/"

adapter_options = {
  fog_provider: "AWS",
  aws_access_key_id: ENV.fetch("S3_ACCESS_KEY_ID"),
  aws_secret_access_key: ENV.fetch("S3_SECRET_ACCESS_KEY"),
  fog_directory: S3_BUCKET_NAME,
  fog_region: S3_REGION
}
SitemapGenerator::Sitemap.adapter = SitemapGenerator::S3Adapter.new(adapter_options)

SitemapGenerator::Sitemap.create do
  add about_path, priority: 0.7
  add popular_works_path, priority: 0.9
  add privacy_path
  add search_works_path, priority: 0.8
  add terms_path

  User.find_each do |u|
    add user_path(u.username), priority: 0.9, lastmod: u.updated_at

    Status.kind.values.each do |k|
      add user_works_path(u.username, k), priority: 0.8
    end
  end

  Season.find_each do |s|
    add season_works_path(s.slug), priority: 0.9
  end

  Person.published.find_each do |p|
    if p.casts.present?
      add person_path(p), priority: 0.8, lastmod: p.updated_at
    end
  end

  Organization.published.find_each do |o|
    if o.work_organizations.present?
      add organization_path(o), priority: 0.8, lastmod: o.updated_at
    end
  end

  Work.published.find_each do |w|
    add work_path(w.id), priority: 1.0, lastmod: w.updated_at

    w.episodes.published.find_each do |e|
      if e.checkins.present?
        add work_episode_path(w.id, e.id), priority: 1.0
      end
    end
  end
end

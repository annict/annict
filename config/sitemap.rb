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
  add db_activities_path, priority: 0.6
  add db_edit_requests_path, priority: 0.6
  add db_root_path, priority: 0.6
  add new_user_registration_path, priority: 0.7
  add new_user_session_path, priority: 0.7
  add popular_works_path, priority: 0.9
  add privacy_path
  add resourceless_db_works_path
  add search_works_path, priority: 0.8
  add terms_path

  EditRequest.find_each do |er|
    add db_edit_request_path(er.id), priority: 0.7
  end

  User.find_each do |u|
    add following_user_path(u.username), priority: 0.8
    add followers_user_path(u.username), priority: 0.8
    add user_path(u.username), priority: 0.9, lastmod: u.updated_at
  end

  Season.find_each do |s|
    add season_works_path(s.slug), priority: 0.9
  end

  Work.find_each do |w|
    add work_path(w.id), priority: 1.0, lastmod: w.updated_at

    w.episodes.find_each do |e|
      add work_episode_path(w.id, e.id), priority: 1.0

      e.checkins.find_each do |c|
        add work_episode_checkin_path(w.id, e.id, c.id), priority: 0.8
      end
    end
  end
end

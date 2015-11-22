# Set the host name for URL creation
SitemapGenerator::Sitemap.default_host = "https://annict.com"

SitemapGenerator::Sitemap.create do
  # Put links creation logic here.
  #
  # The root path '/' and sitemap index file are added automatically for you.
  # Links are added to the Sitemap in the order they are specified.
  #
  # Usage: add(path, options={})
  #        (default options are used if you don't specify)
  #
  # Defaults: :priority => 0.5, :changefreq => 'weekly',
  #           :lastmod => Time.now, :host => default_host
  #
  # Examples:
  #
  # Add '/articles'
  #
  #   add articles_path, :priority => 0.7, :changefreq => 'daily'
  #
  # Add all articles:
  #
  #   Article.find_each do |article|
  #     add article_path(article), :lastmod => article.updated_at
  #   end

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
    add user_path(u.username), priority: 0.9
  end

  Season.find_each do |s|
    add season_works_path(s.slug), priority: 0.9
  end

  Work.find_each do |w|
    add work_path(w.id), priority: 1.0

    w.episodes.find_each do |e|
      add work_episode_path(w.id, e.id), priority: 1.0

      e.checkins.find_each do |c|
        add work_episode_checkin_path(w.id, e.id, c.id), priority: 0.8
      end
    end
  end
end

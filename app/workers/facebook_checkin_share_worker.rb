class FacebookCheckinShareWorker
  include Sidekiq::Worker

  def perform(checkin_id)
    checkin = Checkin.find(checkin_id)

    if checkin.facebook_url_hash.blank?
      checkin.update_column(:facebook_url_hash, checkin.generate_url_hash)
    end

    facebook = checkin.user.providers.where(name: 'facebook').first
    graph = Koala::Facebook::API.new(facebook.token)

    work_title = checkin.episode.work.title
    episode_number_title = checkin.episode.number_title
    title = "#{work_title} #{episode_number_title}".truncate(30)

    message = checkin.comment.squish.truncate(50) if checkin.comment.present?
    link = if Rails.env.development?
             "http://www.annict.com/checkins/redirect/fb/#{checkin.facebook_url_hash}"
           else
             "http://#{ENV['HOST']}/checkins/redirect/fb/#{checkin.facebook_url_hash}"
           end
    name = I18n.t('checkins.facebook_share_text', title: title)
    caption = "Checkined by #{checkin.user.username}"

    graph.put_connections('me', 'feed',
      message:      message,
      link:         link,
      name:         name,
      caption:      caption
    )
  end
end
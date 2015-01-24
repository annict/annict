require 'open-uri'

class TwitterWatchingShareWorker
  include Sidekiq::Worker

  def perform(user_id, body)
    user = User.find(user_id)
    shot = user.shots.last
    image_url = open("#{ENV['ANNICT_IMAGE2_URL']}/images/#{shot.image_uid}")
    body = body.present? ? "#{body.truncate(53)} " : ''
    tweet = "#{body}#{ENV['ANNICT_URL']}/users/#{user.username} "

    TwitterService.new(user).client.update_with_media(tweet, image_url)
  end
end

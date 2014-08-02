class SyobocalMailer < ActionMailer::Base
  default from: 'Annict <no-reply@annict.com>'

  def episode_created_notification(episode_id)
    @episode = Episode.find(episode_id)

    mail(to: 'anannict@gmail.com', subject: t('episodes.created_notification'))
  end
end
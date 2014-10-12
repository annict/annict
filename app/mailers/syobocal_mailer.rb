class SyobocalMailer < ActionMailer::Base
  default from: 'Annict <no-reply@annict.com>'

  def episode_created_notification(episode_id)
    @episode = Episode.find(episode_id)

    mail(to: 'anannict@gmail.com', subject: t('episodes.created_notification'))
  end

  def special_program_notification(alert_id)
    alert = Syobocal::Alert.find(alert_id)
    @work = alert.work
    @sub_title = alert.sc_sub_title
    @prog_comment = alert.sc_prog_comment

    mail(to: 'anannict@gmail.com', subject: t('syobocal.special_program_notification'))
  end
end

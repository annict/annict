class AppealMailer < ActionMailer::Base
  default from: 'Annict <no-reply@annict.com>'

  def update_request(user_id, work_id)
    @user = User.find(user_id)
    @work = Work.find(work_id)

    mail(to: 'anannict@gmail.com', subject: t('appeals.receive_update_request'))
  end
end
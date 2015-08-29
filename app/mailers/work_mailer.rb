class WorkMailer < ActionMailer::Base
  default from: "Annict <no-reply@annict.com>"

  def untouched_works_notification(work_ids)
    @works = Work.where(id: work_ids)

    mail(to: "anannict@gmail.com", subject: "【Annict DB】未更新の作品を更新して下さい")
  end
end

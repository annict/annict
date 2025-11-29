# typed: false
# frozen_string_literal: true

class WorkMailer < ApplicationMailer
  def untouched_works_notification(work_ids)
    @works = Work.where(id: work_ids)

    mail(to: "me@shimba.co", subject: "【Annict DB】未更新の作品を更新して下さい")
  end
end

# frozen_string_literal: true

class WorkMailer < ApplicationMailer
  def untouched_works_notification(work_ids)
    @works = Anime.where(id: work_ids)

    mail(to: "hello@annict.com", subject: "【Annict DB】未更新の作品を更新して下さい")
  end
end

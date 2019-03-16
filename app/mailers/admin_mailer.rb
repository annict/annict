# frozen_string_literal: true

class AdminMailer < ActionMailer::Base
  default from: "Annict <no-reply@annict.com>"

  def episode_created_notification(episode_id)
    @episode = Episode.find(episode_id)

    mail(to: "shimbaco@annict.com", subject: "エピソードが追加されました")
  end
end

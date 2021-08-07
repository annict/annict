# frozen_string_literal: true

class AdminMailer < ApplicationMailer
  def episode_created_notification(episode_id)
    @episode = Episode.find(episode_id)

    mail(to: "hello@annict.com", subject: "エピソードが追加されました")
  end

  def error_in_episode_generator_notification(slot_id, error_message)
    @slot = Slot.find(slot_id)
    @work = @slot.anime
    @error_message = error_message

    mail(to: "hello@annict.com", subject: "エピソード生成中にエラーが発生しました")
  end
end

# typed: false
# frozen_string_literal: true

class ForumMailer < ApplicationMailer
  def comment_notification(user_id, comment_id)
    @receiver = User.find(user_id)
    @comment = ForumComment.find(comment_id)
    @sender = @comment.user
    @post_title = @comment.forum_post.title
    @username = @sender.username
    @name = @sender.profile.name

    I18n.with_locale(@receiver.locale) do
      subject = default_i18n_subject(
        name: @name,
        username: @username,
        post_title: @post_title
      )
      mail(to: @receiver.email, subject: subject)
    end
  end
end

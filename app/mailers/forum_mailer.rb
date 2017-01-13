# frozen_string_literal: true

class ForumMailer < ActionMailer::Base
  default from: "Annict <no-reply@annict.com>"

  def comment_notification(comment_id, email)
    @comment = ForumComment.find(comment_id)
    @username = @comment.user.username
    @name = @comment.user.profile.name

    subject = t "messages.forum.comments.comment_notification_subject",
      name: @name,
      username: @username
    mail(to: email, subject: subject)
  end
end

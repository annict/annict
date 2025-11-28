# typed: false
# frozen_string_literal: true

class MentionMailer < ApplicationMailer
  def notify(username, resource_id, resource_type, column)
    @user = User.where(username: username).first
    return if @user.blank?

    @resource = resource_type.constantize.find(resource_id)
    @sender = @resource.user
    @body = @resource.send(column)
    @reply_path = reply_path(@resource)

    subject = default_i18n_subject(name: @sender.profile.name, username: @sender.username)
    mail(to: @user.email, subject: subject)
  end

  private

  def reply_path(resource)
    case resource.class.name
    when "DbComment"
      "/db/#{resource.resource_type.tableize}/#{resource.resource_id}/activities"
    end
  end
end

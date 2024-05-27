# typed: false
# frozen_string_literal: true

module Mentionable
  extend ActiveSupport::Concern

  included do
    def notify_mentioned_users(*columns)
      columns.each do |column|
        usernames = send(column).scan(/@[A-Za-z0-9_]+/).map { |str|
          str[0] = ""
          str
        }

        usernames.each do |username|
          MentionMailer.notify(username, id, self.class.name, column.to_s).deliver_later
        end
      end
    end
  end
end

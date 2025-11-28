# typed: false

class Notification < ApplicationRecord
  counter_culture :user

  belongs_to :action_user, class_name: "User"
  belongs_to :trackable, polymorphic: true
  belongs_to :user

  scope :unread, -> { where(read: false) }
end

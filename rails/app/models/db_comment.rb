# typed: false
# frozen_string_literal: true

class DbComment < ApplicationRecord
  include Mentionable
  include UgcLocalizable

  belongs_to :resource, polymorphic: true
  belongs_to :user

  validates :body, presence: true

  after_create :create_db_activity!
  after_commit -> { notify_mentioned_users(:body) }

  def create_db_activity!
    DbActivity.create! do |a|
      a.user = user
      a.root_resource = resource.root_resource
      a.trackable = resource
      a.object = self
      a.action = "comments.create"
    end
  end

  def anchor
    "#{self.class.name}#{id}"
  end
end

# frozen_string_literal: true
# == Schema Information
#
# Table name: db_comments
#
#  id            :integer          not null, primary key
#  user_id       :integer          not null
#  resource_id   :integer          not null
#  resource_type :string           not null
#  body          :text             not null
#  created_at    :datetime         not null
#  updated_at    :datetime         not null
#
# Indexes
#
#  index_db_comments_on_resource_id_and_resource_type  (resource_id,resource_type)
#  index_db_comments_on_user_id                        (user_id)
#

class DbComment < ApplicationRecord
  include Mentionable

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

# == Schema Information
#
# Table name: edit_request_comments
#
#  id              :integer          not null, primary key
#  edit_request_id :integer          not null
#  user_id         :integer          not null
#  body            :text             not null
#  created_at      :datetime         not null
#  updated_at      :datetime         not null
#
# Indexes
#
#  index_edit_request_comments_on_edit_request_id  (edit_request_id)
#  index_edit_request_comments_on_user_id          (user_id)
#

class EditRequestComment < ActiveRecord::Base
  belongs_to :edit_request
  belongs_to :user
  has_many :db_activities, as: :trackable, dependent: :destroy

  validates :body, presence: true

  after_create :create_participant
  after_create :create_db_activity

  private

  def create_participant
    edit_request.participants.where(user: user).first_or_create
  end

  def create_db_activity
    DbActivity.create do |a|
      a.user = user
      a.recipient = edit_request
      a.trackable = self
      a.action = "edit_request_comments.create"
    end
  end
end

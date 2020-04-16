# == Schema Information
#
# Table name: notifications
#
#  id             :bigint           not null, primary key
#  action         :string(510)      not null
#  read           :boolean          default(FALSE), not null
#  trackable_type :string(510)      not null
#  created_at     :datetime
#  updated_at     :datetime
#  action_user_id :bigint           not null
#  trackable_id   :bigint           not null
#  user_id        :bigint           not null
#
# Indexes
#
#  notifications_action_user_id_idx  (action_user_id)
#  notifications_user_id_idx         (user_id)
#
# Foreign Keys
#
#  notifications_action_user_id_fk  (action_user_id => users.id) ON DELETE => cascade
#  notifications_user_id_fk         (user_id => users.id) ON DELETE => cascade
#

class Notification < ApplicationRecord
  belongs_to :action_user, class_name: 'User'
  belongs_to :trackable,   polymorphic: true
  belongs_to :user

  scope :unread, -> { where(read: false) }

  after_create :increment_notifications_count
  after_destroy :decrement_notifications_count


  private

  # 通知ページにアクセスしたとき、通知数をリセットするので、
  # read onlyになるカウンターキャッシュを使用せず、自前でインクリメントする
  def increment_notifications_count
    user.increment!(:notifications_count)
  end

  def decrement_notifications_count
    user.decrement!(:notifications_count) if user.notifications_count > 0
  end
end

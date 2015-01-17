# == Schema Information
#
# Table name: notifications
#
#  id             :integer          not null, primary key
#  user_id        :integer          not null
#  action_user_id :integer          not null
#  trackable_id   :integer          not null
#  trackable_type :string(510)      not null
#  action         :string(510)      not null
#  read           :boolean          default("false"), not null
#  created_at     :datetime
#  updated_at     :datetime
#
# Indexes
#
#  notifications_action_user_id_idx  (action_user_id)
#  notifications_user_id_idx         (user_id)
#

class Notification < ActiveRecord::Base
  belongs_to :action_user, class_name: 'User'
  belongs_to :trackable,   polymorphic: true
  belongs_to :user

  scope :unread, -> { where(read: false) }

  after_create  :increment_notifications_count
  after_destroy :decrement_notifications_count


  private

  # 通知ページにアクセスしたとき、通知数をリセットするので、
  # read onlyになるカウンターキャッシュを使用せず、自前でインクリメントする
  def increment_notifications_count
    user.increment!(:notifications_count)
  end

  def decrement_notifications_count
    user.decrement!(:notifications_count)
  end
end

# == Schema Information
#
# Table name: notifications
#
#  id             :integer          not null, primary key
#  user_id        :integer          not null
#  action_user_id :integer          not null
#  trackable_id   :integer          not null
#  trackable_type :string(255)      not null
#  action         :string(255)      not null
#  read           :boolean          default(FALSE), not null
#  created_at     :datetime
#  updated_at     :datetime
#
# Indexes
#
#  index_notifications_on_read                             (read)
#  index_notifications_on_trackable_id_and_trackable_type  (trackable_id,trackable_type)
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

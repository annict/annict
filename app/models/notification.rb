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
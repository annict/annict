# frozen_string_literal: true

class UserWorkRecordsQuery
  def call(work_records:, user:)
    @work_records = work_records
    @user = user

    @work_records = @work_records.
      without_deleted.
      preload(:record, user: :profile).
      joins(:user).
      merge(User.without_deleted)
    @work_records = @work_records.where.not(user_id: @user.mute_users.pluck(:muted_user_id)) if @user
    @work_records = join_likes if @user
    @work_records = join_latest_statuses if @user

    selects = ["work_records.*"]
    selects << "likes.id AS user_like_id" if @user
    selects << "latest_statuses.id AS user_latest_status_id" if @user
    @work_records.select(selects.join(", "))
  end

  private

  def join_likes
    sql = [
      "
        LEFT OUTER JOIN likes ON
          likes.recipient_type = 'WorkRecord' AND
          work_records.id = likes.recipient_id AND
          likes.user_id = %s
      ", @user.id
    ]
    @work_records.joins(Like.sanitize_sql_array(sql))
  end

  def join_latest_statuses
    sql = [
      "
        LEFT OUTER JOIN latest_statuses ON
          work_records.work_id = latest_statuses.work_id AND
          latest_statuses.kind IN (1, 2, 5) AND
          latest_statuses.user_id = %s
      ", @user.id
    ]
    @work_records.joins(LatestStatus.sanitize_sql_array(sql))
  end
end

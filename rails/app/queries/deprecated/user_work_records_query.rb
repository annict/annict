# typed: false
# frozen_string_literal: true

class Deprecated::UserWorkRecordsQuery
  def call(work_records:, user:)
    @work_records = work_records
    @user = user

    @work_records = @work_records
      .only_kept
      .preload(:record, user: :profile)
      .joins(:user)
      .merge(User.only_kept)
    @work_records = @work_records.where.not(user_id: @user.mute_users.pluck(:muted_user_id)) if @user
    @work_records = join_likes if @user
    @work_records = join_library_entries if @user

    selects = ["work_records.*"]
    selects << "likes.id AS user_like_id" if @user
    selects << "library_entries.id AS user_library_entry_id" if @user
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

  def join_library_entries
    sql = [
      "
        LEFT OUTER JOIN library_entries
          INNER JOIN statuses ON statuses.id = library_entries.status_id ON
            work_records.work_id = library_entries.work_id AND
            statuses.kind IN (1, 2, 5) AND
            library_entries.user_id = %s
      ", @user.id
    ]
    @work_records.joins(LibraryEntry.sanitize_sql_array(sql))
  end
end

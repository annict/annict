# frozen_string_literal: true

class UserActivitiesQuery
  def call(activities:, user:, page:)
    @activities = activities
    @user = user

    if @user
      @activities = join_likes
      @activities = join_latest_statuses
      @activities = join_work_records
      @activities = join_episode_records
    end

    selects = ["activities.*"]
    if @user
      selects << "likes.id AS user_like_id"
      selects << "latest_statuses.id AS user_latest_status_id"
      selects << "work_records.id AS user_work_record_id"
      selects << "episode_records.id AS user_episode_record_id"
    end

    @activities.
      order(id: :desc).
      preload(
        :work_record,
        :episode_record,
        :multiple_episode_record,
        :status,
        work: :work_image,
        episode: { work: :work_image },
        user: :profile
      ).
      joins(:work).
      merge(Work.published).
      select(selects.join(", ")).
      page(page)
  end

  private

  def join_likes
    sql = [
      "
        LEFT OUTER JOIN likes ON
          activities.trackable_type = likes.recipient_type AND
          activities.trackable_id = likes.recipient_id AND
          likes.user_id = %s
      ", @user.id
    ]
    @activities.joins(Like.sanitize_sql_array(sql))
  end

  def join_latest_statuses
    sql = [
      "
        LEFT OUTER JOIN latest_statuses ON
          (
            activities.action = 'create_work_record' OR
            activities.action = 'create_episode_record'
          ) AND
          activities.work_id = latest_statuses.work_id AND
          latest_statuses.kind IN (1, 2, 5) AND
          latest_statuses.user_id = %s
      ", @user.id
    ]
    @activities.joins(LatestStatus.sanitize_sql_array(sql))
  end

  def join_work_records
    sql = [
      "
        LEFT OUTER JOIN work_records ON
          activities.action = 'create_work_record' AND
          activities.work_id = work_records.work_id AND
          work_records.user_id = %s
      ", @user.id
    ]
    @activities.joins(WorkRecord.sanitize_sql_array(sql))
  end

  def join_episode_records
    sql = [
      "
        LEFT OUTER JOIN episode_records ON
          activities.action = 'create_episode_record' AND
          activities.episode_id = episode_records.episode_id AND
          episode_records.user_id = %s
      ", @user.id
    ]
    @activities.joins(EpisodeRecord.sanitize_sql_array(sql))
  end
end

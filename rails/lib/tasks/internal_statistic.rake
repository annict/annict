# typed: false
# frozen_string_literal: true

namespace :internal_statistic do
  task calc: :environment do
    today = Date.today
    users_past_week = User.past_week
    users_yesterday = User.yesterday

    [
      {key_name: "episode_records", model: Record},
      {key_name: "statuses", model: Status},
      {key_name: "work_records", model: WorkRecord}
    ].each do |h|
      InternalStatistic.where(key: "users_count_created_#{h[:key_name]}_in_all", date: today).first_or_create! do |is|
        is.value = h[:model].select("user_id").group(:user_id).length
      end

      InternalStatistic.where(key: "users_count_created_#{h[:key_name]}_in_past_week", date: today).first_or_create! do |is|
        is.value = h[:model].select("user_id, MAX(created_at)").group(:user_id).past_week.length
      end

      InternalStatistic.where(key: "users_count_created_#{h[:key_name]}_in_yesterday", date: today).first_or_create! do |is|
        is.value = h[:model].select("user_id, MAX(created_at)").group(:user_id).yesterday.length
      end
    end

    InternalStatistic.where(key: :users_count_registered_in_all, date: today).first_or_create! do |is|
      is.value = User.count
    end

    InternalStatistic.where(key: :users_count_registered_in_yesterday, date: today).first_or_create! do |is|
      is.value = users_yesterday.count
    end

    InternalStatistic.where(key: :users_count_registered_in_past_week, date: today).first_or_create! do |is|
      is.value = users_past_week.count
    end

    InternalStatistic.where(key: :users_count_active_in_all_users_past_week, date: today).first_or_create! do |is|
      is.value = Activity.select("user_id, MAX(created_at)").group(:user_id).past_week.length
    end

    InternalStatistic.where(key: :users_count_active_in_new_users_past_week, date: today).first_or_create! do |is|
      is.value = Activity
        .where(user_id: users_past_week.pluck(:id))
        .select("user_id, MAX(created_at)")
        .group(:user_id)
        .past_week
        .length
    end

    InternalStatisticMailer.result_mail(today.to_s).deliver_later
  end
end

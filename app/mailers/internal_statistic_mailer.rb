# frozen_string_literal: true

class InternalStatisticMailer < ActionMailer::Base
  default from: "Annict <no-reply@annict.com>"

  def result_mail(date_str)
    statistcs = InternalStatistic.where(date: Date.parse(date_str))
    @users_count_registered_in_all = statistcs.
      find_by(key: :users_count_registered_in_all)&.value.presence || 0
    @users_count_registered_in_new = statistcs.
      find_by(key: :users_count_registered_in_new)&.value.presence || 0
    @users_count_active_in_all_users_past_week = statistcs.
      find_by(key: :users_count_active_in_all_users_past_week)&.value.presence || 0
    @users_count_active_in_new_users_past_week = statistcs.
      find_by(key: :users_count_active_in_new_users_past_week)&.value.presence || 0

    mail(to: "anannict@gmail.com", subject: "Annict Statistic - #{date_str}")
  end
end

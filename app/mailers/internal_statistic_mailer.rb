# frozen_string_literal: true

class InternalStatisticMailer < ActionMailer::Base
  default from: "Annict <no-reply@annict.com>"

  def result_mail(date_str)
    statistcs = InternalStatistic.where(date: Date.parse(date_str))
    @data = statistcs.map do |s|
      [s.key, s.value.presence || 0]
    end
    @data = Hash[@data]

    mail(to: "hello@annict.com", subject: "Annict Statistic - #{date_str}")
  end
end

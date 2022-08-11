# frozen_string_literal: true

class InternalStatisticMailer < ApplicationMailer
  def result_mail(date_str)
    statistcs = InternalStatistic.where(date: Date.parse(date_str))
    @data = statistcs.map { |s|
      [s.key, s.value.presence || 0]
    }
    @data = @data.to_h

    mail(to: "me@shimba.co", subject: "Annict Statistic - #{date_str}")
  end
end

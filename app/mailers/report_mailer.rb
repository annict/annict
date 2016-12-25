# frozen_string_literal: true

class ReportMailer < ActionMailer::Base
  default from: "Annict <no-reply@annict.com>"

  def notify(report_id)
    @report = Report.find(report_id)

    mail(to: "admin@annict.com", subject: "[Annict] Report received")
  end
end

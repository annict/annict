# typed: strict
# frozen_string_literal: true

require "icalendar/tzinfo"

module IcalendarHelper
  extend T::Sig

  sig { params(user: User, slots: ActiveRecord::Relation, works: ActiveRecord::Relation).returns(String) }
  def render_user_calendar(user:, slots:, works:)
    cal = Icalendar::Calendar.new
    tz_identifier = user.time_zone

    cal.append_custom_property("X-WR-TIMEZONE", tz_identifier)
    cal.append_custom_property("X-WR-CALNAME", "Annict@#{user.username}")

    tz = TZInfo::Timezone.get(tz_identifier)
    timezone = tz.ical_timezone(DateTime.now)
    cal.add_timezone(timezone)

    slots.each do |s|
      cal.event do |e|
        e.dtstart = Icalendar::Values::DateTime.new(s.started_at.in_time_zone(tz_identifier).strftime(Icalendar::Values::DateTime::FORMAT))
        e.dtend = Icalendar::Values::DateTime.new((s.started_at + 30.minutes).in_time_zone(tz_identifier).strftime(Icalendar::Values::DateTime::FORMAT))
        e.summary = Icalendar::Values::Text.new("#{s.work.local_title} #{s.episode.title_with_number} (#{s.channel.name})")
        e.description = Icalendar::Values::Text.new("#{s.work.local_title} #{s.episode.title_with_number}\n#{episode_url(s.work_id, s.episode_id)}")
      end
    end

    works.each do |w|
      cal.event do |e|
        e.dtstart = Icalendar::Values::Date.new(w.started_on.strftime("%Y%m%d"))
        e.dtend = Icalendar::Values::Date.new((w.started_on + 1.day).strftime("%Y%m%d"))
        e.summary = w.local_title
        e.description = "#{w.local_title} #{work_url(w)}"
      end
    end

    cal.publish

    cal.to_ical
  end
end

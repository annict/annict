# frozen_string_literal: true

module TimeZoneHelper
  def display_time(time)
    time = time&.in_time_zone(local_time_zone)
    return time&.strftime("%Y-%m-%d %H:%M") if locale_ja?
    time&.strftime("%b %-d, %Y %H:%M")
  end

  def display_date(date)
    return date.strftime("%Y-%m-%d") if locale_ja?
    date.strftime("%b %-d, %Y")
  end

  def decorated_tz_name(time_zone)
    offset = time_zone.utc_offset / 3600
    formatted_offset = format("%02d:00", offset.abs)
    formatted_offset = offset >= 0 ? "+#{formatted_offset}" : "-#{formatted_offset}"
    "(GMT#{formatted_offset}) #{time_zone.name}"
  end

  def local_time_zone
    current_user&.time_zone.presence || cookies["ann_time_zone"]
  end
end

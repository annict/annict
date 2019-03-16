# frozen_string_literal: true

namespace :tmp do
  task create_program_details: :environment do
    Work.published.find_each do |w|
      next if w.programs.empty?

      puts "--- work: #{w.id}"

      programs = w.programs.published.from("
        (
          select *,
            row_number() over (
              partition by channel_id
              order by started_at asc
            ) as row_num
          from programs
          where programs.work_id = #{w.id} and programs.aasm_state = 'published'
        ) as programs
      ").where("programs.row_num <= 1")

      programs.each do |p|
        puts "--- work: #{w.id} program: #{p.id}"

        w.
          program_details.
          published.
          where(
            channel_id: p.channel_id,
            started_at: p.started_at,
            rebroadcast: p.rebroadcast
          ).
          first_or_create!
      end
    end
  end
end

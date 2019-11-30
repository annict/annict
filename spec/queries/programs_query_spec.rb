# frozen_string_literal: true

describe ProgramsQuery, type: :query do
  context "when the `order` option is not specified" do
    let!(:program_1) { create :program, created_at: Time.new(2019, 12, 1), started_at: Time.new(2019, 12, 31) }
    let!(:program_2) { create :program, created_at: Time.new(2019, 12, 2), started_at: Time.new(2019, 12, 30) }
    let!(:program_3) { create :program, created_at: Time.new(2019, 12, 3), started_at: Time.new(2019, 12, 29) }

    it "returns programs which are sorted to asc by `created_at` field" do
      programs = ProgramsQuery.new(
        Program.all
      ).call

      expect(programs.pluck(:id)).to match([program_1.id, program_2.id, program_3.id])
    end
  end

  context "when the `order` option is specified" do
    let!(:program_1) { create :program, created_at: Time.new(2019, 12, 1), started_at: Time.new(2019, 12, 31) }
    let!(:program_2) { create :program, created_at: Time.new(2019, 12, 2), started_at: Time.new(2019, 12, 30) }
    let!(:program_3) { create :program, created_at: Time.new(2019, 12, 3), started_at: Time.new(2019, 12, 29) }

    it "returns programs which are sorted by specified field" do
      programs = ProgramsQuery.new(
        Program.all,
        order: OrderProperty.new(:started_at, :asc)
      ).call

      expect(programs.pluck(:id)).to match([program_3.id, program_2.id, program_1.id])
    end
  end
end

# frozen_string_literal: true

describe UserSlotsQuery, type: :query do
  let!(:user) { create :user }
  let!(:setting) { create :setting, user: user }
  let!(:work_1) { create :work }
  let!(:work_2) { create :work }
  let!(:work_3) { create :work }
  let!(:episode_1) { create :episode, work: work_1 }
  let!(:episode_2) { create :episode, work: work_1 }
  let!(:episode_3) { create :episode, work: work_2 }
  let!(:episode_4) { create :episode, work: work_2 }
  let!(:episode_5) { create :episode, work: work_3 }
  let!(:channel_1) { create :channel }
  let!(:channel_2) { create :channel }
  let!(:program_1) { create :program, channel: channel_1, work: work_1 }
  let!(:program_2) { create :program, channel: channel_1, work: work_2 }
  let!(:program_3) { create :program, channel: channel_2, work: work_2 }
  let!(:program_4) { create :program, channel: channel_1, work: work_1, rebroadcast: true }
  let!(:slot_1) { create :slot, program: program_1, episode: episode_1, started_at: Time.new(2020, 1, 7, 1, 0) }
  let!(:slot_2) { create :slot, program: program_1, episode: episode_2, started_at: Time.new(2020, 1, 14, 1, 0) }
  let!(:slot_3) { create :slot, program: program_2, episode: episode_3, started_at: Time.new(2020, 1, 7, 1, 30) }
  let!(:slot_4) { create :slot, program: program_2, episode: episode_4, started_at: Time.new(2019, 1, 14, 1, 30) }
  let!(:slot_5) { create :slot, program: program_3, episode: episode_3, started_at: Time.new(2019, 1, 8, 23, 0) }
  let!(:slot_6) { create :slot, program: program_3, episode: episode_4, started_at: Time.new(2019, 1, 15, 23, 0) }
  let!(:slot_7) { create :slot, program: program_4, episode: episode_1, started_at: Time.new(2020, 10, 1), rebroadcast: true }
  let!(:slot_8) { create :slot, program: program_4, episode: episode_2, started_at: Time.new(2020, 10, 8), rebroadcast: true }

  context "when the user does not set channel work" do
    context "when the user does not set status on works" do
      context "when the `watched` option is not specified" do
        it "returns no slots" do
          slots = UserSlotsQuery.new(
            user,
            Slot.all
          ).call

          expect(slots.pluck(:id)).to match([])
        end
      end

      context "when the `watched` option is `true`" do
        it "returns no slots" do
          slots = UserSlotsQuery.new(
            user,
            Slot.all,
            watched: true
          ).call

          expect(slots.pluck(:id)).to match([])
        end
      end

      context "when the `watched` option is `false`" do
        it "returns no slots" do
          slots = UserSlotsQuery.new(
            user,
            Slot.all,
            watched: false
          ).call

          expect(slots.pluck(:id)).to match([])
        end
      end
    end

    context "when the user is watching work_1" do
      let!(:status) { create :status, user: user, work: work_1, kind: :watching }

      context "when the user is watching work_2" do
        let!(:status) { create :status, user: user, work: work_2, kind: :watching }

        context "when the user does not watch episodes" do
          context "when the `watched` option is not specified" do
            it "returns no slots" do
              slots = UserSlotsQuery.new(
                user,
                Slot.all
              ).call

              expect(slots.pluck(:id)).to match([])
            end
          end

          context "when the `watched` option is `true`" do
            it "returns no slots" do
              slots = UserSlotsQuery.new(
                user,
                Slot.all,
                watched: true
              ).call

              expect(slots.pluck(:id)).to match([])
            end
          end

          context "when the `watched` option is `false`" do
            it "returns no slots" do
              slots = UserSlotsQuery.new(
                user,
                Slot.all,
                watched: false
              ).call

              expect(slots.pluck(:id)).to match([])
            end
          end
        end

        context "when the user watches episode_1" do
          let!(:episode_record) { build :episode_record, user: user, episode: episode_1 }
          let!(:library_entry) { create(:library_entry, user: user, work: episode_1.work, watched_episode_ids: [episode_1]) }

          context "when the `watched` option is not specified" do
            it "returns no slots" do
              slots = UserSlotsQuery.new(
                user,
                Slot.all
              ).call

              expect(slots.pluck(:id)).to match([])
            end
          end

          context "when the `watched` option is `true`" do
            it "returns no slots" do
              slots = UserSlotsQuery.new(
                user,
                Slot.all,
                watched: true
              ).call

              expect(slots.pluck(:id)).to match([])
            end
          end

          context "when the `watched` option is `false`" do
            it "returns no slots" do
              slots = UserSlotsQuery.new(
                user,
                Slot.all,
                watched: false
              ).call

              expect(slots.pluck(:id)).to match([])
            end
          end
        end
      end
    end
  end

  context "when the user sets channel work" do
    let!(:channel_work_1) { create :channel_work, user: user, channel: channel_1, work: work_1 }
    let!(:channel_work_2) { create :channel_work, user: user, channel: channel_1, work: work_2 }

    context "when the user does not set status on works" do
      context "when the `watched` option is not specified" do
        it "returns no slots" do
          slots = UserSlotsQuery.new(
            user,
            Slot.all
          ).call

          expect(slots.pluck(:id)).to match([])
        end
      end

      context "when the `watched` option is `true`" do
        it "returns no slots" do
          slots = UserSlotsQuery.new(
            user,
            Slot.all,
            watched: true
          ).call

          expect(slots.pluck(:id)).to match([])
        end
      end

      context "when the `watched` option is `false`" do
        it "returns no slots" do
          slots = UserSlotsQuery.new(
            user,
            Slot.all,
            watched: false
          ).call

          expect(slots.pluck(:id)).to match([])
        end
      end
    end

    context "when the user is watching work_1" do
      let!(:status_1) { create :status, user: user, work: work_1, kind: :watching }
      let!(:library_entry_1) { create(:library_entry, user: user, work: work_1, status: status_1) }

      context "when the user is watching work_2" do
        let!(:status_2) { create :status, user: user, work: work_2, kind: :watching }
        let!(:library_entry_2) { create(:library_entry, user: user, work: work_2, status: status_2) }

        context "when the user does not watch episodes" do
          context "when the `watched` option is not specified" do
            it "returns slots" do
              slots = UserSlotsQuery.new(
                user,
                Slot.all
              ).call

              expect(slots.pluck(:id)).to contain_exactly(slot_3.id, slot_4.id, slot_7.id, slot_8.id)
            end
          end

          context "when the `watched` option is `true`" do
            it "returns no slots" do
              slots = UserSlotsQuery.new(
                user,
                Slot.all,
                watched: true
              ).call

              expect(slots.pluck(:id)).to match([])
            end
          end

          context "when the `watched` option is `false`" do
            it "returns slots" do
              slots = UserSlotsQuery.new(
                user,
                Slot.all,
                watched: false
              ).call

              expect(slots.pluck(:id)).to contain_exactly(slot_3.id, slot_4.id, slot_7.id, slot_8.id)
            end
          end
        end

        context "when the user watches episode_1" do
          let!(:episode_record) { create(:episode_record, user: user, episode: episode_1) }

          before do
            library_entry_1.update(watched_episode_ids: [episode_1.id])
          end

          context "when the `watched` option is not specified" do
            it "returns slots" do
              slots = UserSlotsQuery.new(
                user,
                Slot.all
              ).call

              expect(slots.pluck(:id)).to contain_exactly(slot_3.id, slot_4.id, slot_7.id, slot_8.id)
            end
          end

          context "when the `watched` option is `true`" do
            it "returns slots" do
              slots = UserSlotsQuery.new(
                user,
                Slot.all,
                watched: true
              ).call

              expect(slots.pluck(:id)).to contain_exactly(slot_7.id)
            end
          end

          context "when the `watched` option is `false`" do
            it "returns slots" do
              slots = UserSlotsQuery.new(
                user,
                Slot.all,
                watched: false
              ).call

              expect(slots.pluck(:id)).to contain_exactly(slot_3.id, slot_4.id, slot_8.id)
            end
          end
        end
      end
    end
  end
end

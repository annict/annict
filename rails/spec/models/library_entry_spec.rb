# typed: false
# frozen_string_literal: true

describe LibraryEntry, type: :model do
  describe "#append_episode!" do
    let(:user) { create :user }
    let(:existing_work) { create :work }
    let(:work) { create :work }
    let(:episode) { create :episode, work: work, sort_number: 100 }

    before do
      user.library_entries.create!(work: existing_work)
    end

    context "ステータスが設定されていないとき" do
      before do
        @library_entry = user.library_entries.create!(work: work)
      end

      it "エピソードが追加できること" do
        expect(LibraryEntry.count).to eq 2
        expect(@library_entry.position).to eq 2
        expect(@library_entry.watched_episode_ids).to eq []
        expect(@library_entry.next_episode).to be_nil
        expect(@library_entry.next_slot).to be_nil
        expect(@library_entry.program).to be_nil
        expect(@library_entry.status).to be_nil

        @library_entry.append_episode!(episode)

        expect(LibraryEntry.count).to eq 2
        expect(@library_entry.position).to eq 2
        expect(@library_entry.watched_episode_ids).to eq [episode.id]
        expect(@library_entry.next_episode).to be_nil
        expect(@library_entry.next_slot).to be_nil
        expect(@library_entry.program).to be_nil
        expect(@library_entry.status).to be_nil
      end
    end

    context "ステータスが「見てる」のとき" do
      let(:status) { create :status, user: user, work: work, kind: :watching }

      before do
        @library_entry = user.library_entries.create!(work: work, status: status)
      end

      it "エピソードが追加できること" do
        expect(LibraryEntry.count).to eq 2
        expect(@library_entry.position).to eq 2
        expect(@library_entry.watched_episode_ids).to eq []
        expect(@library_entry.next_episode).to be_nil
        expect(@library_entry.next_slot).to be_nil
        expect(@library_entry.program).to be_nil
        expect(@library_entry.status).to eq status

        @library_entry.append_episode!(episode)

        expect(LibraryEntry.count).to eq 2
        # 「見てる」ときは一番上に持って行く
        expect(@library_entry.position).to eq 1
        expect(@library_entry.watched_episode_ids).to eq [episode.id]
        expect(@library_entry.next_episode).to be_nil
        expect(@library_entry.next_slot).to be_nil
        expect(@library_entry.program).to be_nil
        expect(@library_entry.status).to eq status
      end
    end

    context "次のエピソードが存在するとき" do
      let!(:next_episode) { create :episode, work: work, sort_number: 200 }

      before do
        @library_entry = user.library_entries.create!(work: work)
      end

      it "エピソードが追加できること" do
        expect(LibraryEntry.count).to eq 2
        expect(@library_entry.position).to eq 2
        expect(@library_entry.watched_episode_ids).to eq []
        expect(@library_entry.next_episode).to be_nil
        expect(@library_entry.next_slot).to be_nil
        expect(@library_entry.program).to be_nil
        expect(@library_entry.status).to be_nil

        @library_entry.append_episode!(episode)

        expect(LibraryEntry.count).to eq 2
        expect(@library_entry.position).to eq 2
        expect(@library_entry.watched_episode_ids).to eq [episode.id]
        expect(@library_entry.next_episode).to eq next_episode
        expect(@library_entry.next_slot).to be_nil
        expect(@library_entry.program).to be_nil
        expect(@library_entry.status).to be_nil
      end
    end

    context "次の放送予定が存在するとき" do
      let!(:next_episode) { create :episode, work: work, sort_number: 200 }
      let!(:program) { create :program, work: work }
      let!(:slot) { create :slot, program: program, work: work, episode: episode }
      let!(:next_slot) { create :slot, program: program, work: work, episode: next_episode }

      before do
        @library_entry = user.library_entries.create!(work: work, program: program)
      end

      it "エピソードが追加できること" do
        expect(LibraryEntry.count).to eq 2
        expect(@library_entry.position).to eq 2
        expect(@library_entry.watched_episode_ids).to eq []
        expect(@library_entry.next_episode).to be_nil
        expect(@library_entry.next_slot).to be_nil
        expect(@library_entry.program).to eq program
        expect(@library_entry.status).to be_nil

        @library_entry.append_episode!(episode)

        expect(LibraryEntry.count).to eq 2
        expect(@library_entry.position).to eq 2
        expect(@library_entry.watched_episode_ids).to eq [episode.id]
        expect(@library_entry.next_episode).to eq next_episode
        expect(@library_entry.next_slot).to eq next_slot
        expect(@library_entry.program).to eq program
        expect(@library_entry.status).to be_nil
      end
    end
  end
end

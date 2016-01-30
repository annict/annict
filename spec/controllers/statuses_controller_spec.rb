# == Schema Information
#
# Table name: statuses
#
#  id          :integer          not null, primary key
#  user_id     :integer          not null
#  work_id     :integer          not null
#  kind        :integer          not null
#  latest      :boolean          default(FALSE), not null
#  likes_count :integer          default(0), not null
#  created_at  :datetime
#  updated_at  :datetime
#
# Indexes
#
#  statuses_user_id_idx  (user_id)
#  statuses_work_id_idx  (work_id)
#

describe StatusesController do
  let(:user) { create(:registered_user) }
  let(:work) { create(:work) }

  before do
    sign_in user
  end

  describe 'POST select' do
    context '「見てる」に変更したとき' do
      before do
        Status.skip_callback(:create, :after, :finish_tips)
        post :select, work_id: work.id, status_kind: :watching
      end

      it '200が返ること' do
        expect(response.status).to eq(200)
      end

      it 'ステータスが「見てる」になること' do
        expect(user.latest_statuses.find_by(work: work).kind).to eq "watching"
      end
    end

    context '未選択状態に戻したとき' do
      let!(:status) { create(:status, user: user, work: work) }

      before do
        post :select, work_id: work.id, status_kind: :no_select
      end

      it '200を返すこと' do
        expect(response.status).to eq(200)
      end

      it 'ステータスがリセットされること' do
        expect(user.statuses.first.latest).to eq(false)
      end
    end
  end
end

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
        expect(user.statuses.kind_of(work).kind).to eq 'watching'
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

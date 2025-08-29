package ingest

import (
	"context"
	"errors"
	mock_ingest "github.com/rhuandantas/metrika/internal/mocks/ingest"
	mock_repo "github.com/rhuandantas/metrika/internal/mocks/repository"
	"github.com/rhuandantas/metrika/internal/models"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/rhuandantas/metrika/internal/smartblox"
	"github.com/rs/zerolog"
	"go.uber.org/mock/gomock"
)

func TestIngestor(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Ingestor Suite")
	defer GinkgoRecover()
}

var _ = Describe("Ingestor", func() {
	var (
		ctrl        *gomock.Controller
		mockClient  *mock_ingest.MockClient
		mockRepo    *mock_repo.MockRepository
		logger      zerolog.Logger
		eventLogger zerolog.Logger
		ing         *Ingestor
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		mockClient = mock_ingest.NewMockClient(ctrl)
		mockRepo = mock_repo.NewMockRepository(ctrl)
		logger = zerolog.Nop()
		eventLogger = zerolog.Nop()
		ing = New(mockClient, time.Millisecond*1, time.Millisecond*2, logger, eventLogger, mockRepo)
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	It("should return context.Canceled when context is canceled", func() {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		mockRepo.EXPECT().SaveMetrics(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
		err := ing.Run(ctx)
		Expect(err).To(Equal(context.Canceled))
	})
	It("should return error on GetStatus failure", func() {
		mockClient.EXPECT().GetStatus(gomock.Any()).Return(smartblox.Status{}, errors.New("fail"))
		err := ing.process(context.Background())
		Expect(err).To(HaveOccurred())
	})
	It("should return error on GetMetrics failure", func() {
		mockClient.EXPECT().GetStatus(gomock.Any()).Return(smartblox.Status{LastRound: 1}, nil)
		mockRepo.EXPECT().LoadMetrics(gomock.Any()).Return(models.Metrics{}, errors.New("fail"))
		err := ing.process(context.Background())
		Expect(err).To(HaveOccurred())
	})
	It("should get metrics from database and return nil", func() {
		mockClient.EXPECT().GetStatus(gomock.Any()).Return(smartblox.Status{LastRound: 0}, nil)
		mockRepo.EXPECT().LoadMetrics(gomock.Any()).Return(models.Metrics{LastRound: 0}, nil)
		err := ing.process(context.Background())
		Expect(err).To(BeNil())
	})
	It("should return error on getting blocks failure", func() {
		mockClient.EXPECT().GetStatus(gomock.Any()).Return(smartblox.Status{LastRound: 2}, nil)
		mockRepo.EXPECT().LoadMetrics(gomock.Any()).Return(models.Metrics{LastRound: 1}, nil)
		mockClient.EXPECT().GetBlock(gomock.Any(), int64(2)).Return(smartblox.Block{}, errors.New("fail"))
		err := ing.process(context.Background())
		Expect(err).To(HaveOccurred())
	})
	It("should return error updating metrics", func() {
		mockClient.EXPECT().GetStatus(gomock.Any()).Return(smartblox.Status{LastRound: 2}, nil)
		mockRepo.EXPECT().LoadMetrics(gomock.Any()).Return(models.Metrics{LastRound: 1}, nil)
		mockClient.EXPECT().GetBlock(gomock.Any(), int64(2)).Return(smartblox.Block{
			Round: 2,
			Txs: []smartblox.TransactionSig{
				{
					Sig: "mock_sig",
					Tx: smartblox.Transaction{
						Receipient: 1,
						Sender:     2,
						Amount:     1000,
						Type:       transactionType,
					},
				},
				{
					Sig: "mock_sig",
					Tx: smartblox.Transaction{
						Receipient: 3,
						Sender:     4,
						Amount:     100,
						Type:       "avoid",
					},
				},
			},
		}, nil)
		mockRepo.EXPECT().SaveMetrics(gomock.Any(), gomock.Any()).Return(errors.New("fail"))
		err := ing.process(context.Background())
		Expect(err).To(HaveOccurred())
	})
	It("should process rounds and return nil", func() {
		mockClient.EXPECT().GetStatus(gomock.Any()).Return(smartblox.Status{LastRound: 2}, nil)
		mockRepo.EXPECT().LoadMetrics(gomock.Any()).Return(models.Metrics{LastRound: 1}, nil)
		mockClient.EXPECT().GetBlock(gomock.Any(), int64(2)).Return(smartblox.Block{
			Round: 2,
			Txs: []smartblox.TransactionSig{
				{
					Sig: "mock_sig",
					Tx: smartblox.Transaction{
						Receipient: 1,
						Sender:     2,
						Amount:     1000,
						Type:       transactionType,
					},
				},
				{
					Sig: "mock_sig",
					Tx: smartblox.Transaction{
						Receipient: 3,
						Sender:     4,
						Amount:     100,
						Type:       "avoid",
					},
				},
			},
		}, nil)
		mockRepo.EXPECT().SaveMetrics(gomock.Any(), gomock.Any()).Return(nil)
		err := ing.process(context.Background())
		Expect(err).To(BeNil())
	})
})

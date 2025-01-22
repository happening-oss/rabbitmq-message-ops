package managers_test

import (
	"context"
	"errors"
	"log/slog"
	"sync/atomic"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/mock"

	"github.com/happening-oss/rabbitmq-message-ops/internal/messaging/mocks"
	rmocks "github.com/happening-oss/rabbitmq-message-ops/internal/messaging/rabbitmq/mocks"
	smocks "github.com/happening-oss/rabbitmq-message-ops/internal/messaging/selectors/mocks"
	"github.com/happening-oss/rabbitmq-message-ops/internal/tests/stubs"
	"github.com/happening-oss/rabbitmq-message-ops/internal/tests/util"

	hmocks "github.com/happening-oss/rabbitmq-message-ops/internal/messaging/management/handlers/mocks"
	"github.com/happening-oss/rabbitmq-message-ops/internal/messaging/management/managers"
)

var _ = Describe("Queue manager", func() {
	var conMock *mocks.Consumer
	var pubMock *mocks.Publisher
	var log *slog.Logger
	var handler *hmocks.MessageHandler
	var selectorMock *smocks.Selector

	var manager managers.Manager

	var ackMock *rmocks.Acknowledger

	var sequenceNumber atomic.Uint64

	BeforeEach(func() {
		conMock = mocks.NewConsumer(GinkgoT())
		pubMock = mocks.NewPublisher(GinkgoT())
		log = slog.New(stubs.NewHandler())
		handler = hmocks.NewMessageHandler(GinkgoT())
		selectorMock = smocks.NewSelector(GinkgoT())

		manager = managers.NewQueueManager(conMock, log, handler, pubMock, selectorMock, "tempQueue")
		ackMock = rmocks.NewAcknowledger(GinkgoT())
	})

	When("messages present", func() {

		var srcMessages []amqp091.Delivery

		BeforeEach(func() {
			srcMessages = []amqp091.Delivery{
				{DeliveryTag: sequenceNumber.Add(1), Acknowledger: ackMock},
				{DeliveryTag: sequenceNumber.Add(1), Acknowledger: ackMock},
				{DeliveryTag: sequenceNumber.Add(1), Acknowledger: ackMock},
			}
			srcQueue := initReadChannel(srcMessages)
			conMock.On(util.NameOf(conMock.Consume), "srcQueue").Return(srcQueue, nil).Once()
		})

		When("selector throws error", func() {
			BeforeEach(func() {
				selectorMock.On(util.NameOf(selectorMock.IsSelected), mock.Anything).Return(false, errors.New("")).Once()
			})

			When("acknowledger fails to reject delivery", func() {
				BeforeEach(func() {
					ackMock.On(util.NameOf(ackMock.Reject), srcMessages[0].DeliveryTag, true).Return(errors.New(""))
				})

				It("returns error", func() {
					err := manager.Manage(context.Background(), "srcQueue")
					Expect(err).To(HaveOccurred())
					for _, msg := range srcMessages {
						Expect(ackMock.AckedTags[msg.DeliveryTag]).ToNot(BeTrue())
					}
				})
			})

			When("acknowledger succeeds to reject delivery", func() {
				BeforeEach(func() {
					ackMock.On(util.NameOf(ackMock.Reject), srcMessages[0].DeliveryTag, true).Return(nil)
				})

				It("returns error", func() {
					err := manager.Manage(context.Background(), "srcQueue")
					Expect(err).To(HaveOccurred())
					for _, msg := range srcMessages {
						Expect(ackMock.AckedTags[msg.DeliveryTag]).ToNot(BeTrue())
					}
				})
			})
		})

		When("selector succeeds but doesn't select message", func() {
			BeforeEach(func() {
				selectorMock.On(util.NameOf(selectorMock.IsSelected), mock.Anything).Return(false, nil).Times(len(srcMessages))
				pubMock.On(util.NameOf(pubMock.Publish), "tempQueue", mock.Anything).Return(nil).Times(len(srcMessages))
				ackMock.On(util.NameOf(ackMock.Ack), mock.Anything, false).Return(nil).Times(len(srcMessages))
				tempQueue := make(<-chan amqp091.Delivery, len(srcMessages))
				conMock.On(util.NameOf(conMock.Consume), "tempQueue").Return(tempQueue, nil).Once()
			})

			It("doesn't call handler", func() {
				err := manager.Manage(context.Background(), "srcQueue")
				Expect(err).ToNot(HaveOccurred())
			})
		})

		When("selector succeeds and selects message", func() {
			BeforeEach(func() {
				selectorMock.On(util.NameOf(selectorMock.IsSelected), mock.Anything).Return(true, nil).Times(len(srcMessages))
			})

			When("handler throws error", func() {
				BeforeEach(func() {
					selectorMock.On(util.NameOf(selectorMock.IsSelected), mock.Anything).Unset() // because of error while handling msg, we don't want to call isSelected N times
					selectorMock.On(util.NameOf(selectorMock.IsSelected), mock.Anything).Return(true, nil).Once()

					handler.On(util.NameOf(handler.Handle), mock.Anything).Return(false, errors.New("")).Once()
				})

				When("acknowledger fails to reject delivery", func() {
					BeforeEach(func() {
						ackMock.On(util.NameOf(ackMock.Reject), srcMessages[0].DeliveryTag, true).Return(errors.New(""))
					})

					It("returns error", func() {
						err := manager.Manage(context.Background(), "srcQueue")
						Expect(err).To(HaveOccurred())
						for _, msg := range srcMessages {
							Expect(ackMock.AckedTags[msg.DeliveryTag]).ToNot(BeTrue())
						}
					})
				})

				When("acknowledger succeeds to reject delivery", func() {
					BeforeEach(func() {
						ackMock.On(util.NameOf(ackMock.Reject), srcMessages[0].DeliveryTag, true).Return(nil)
					})

					It("returns error", func() {
						err := manager.Manage(context.Background(), "srcQueue")
						Expect(err).To(HaveOccurred())
						for _, msg := range srcMessages {
							Expect(ackMock.AckedTags[msg.DeliveryTag]).ToNot(BeTrue())
						}
					})
				})
			})

			When("handler succeeds but doesn't requeue messages", func() {
				BeforeEach(func() {
					handler.On(util.NameOf(handler.Handle), mock.Anything).Return(false, nil).Times(len(srcMessages))
					for _, msg := range srcMessages {
						ackMock.On(util.NameOf(ackMock.Ack), msg.DeliveryTag, false).Return(nil)
					}
					tempQueue := make(<-chan amqp091.Delivery, len(srcMessages))
					conMock.On(util.NameOf(conMock.Consume), "tempQueue").Return(tempQueue, nil).Once()
				})

				It("acknowledges all source queue messages and succeeds without publishing", func() {
					err := manager.Manage(context.Background(), "srcQueue")
					Expect(err).ToNot(HaveOccurred())
					for _, msg := range srcMessages {
						Expect(ackMock.AckedTags[msg.DeliveryTag]).To(BeTrue())
					}
				})
			})

			When("handler succeeds and requeues message", func() {
				BeforeEach(func() {
					handler.On(util.NameOf(handler.Handle), mock.Anything).Return(true, nil).Times(len(srcMessages))
				})

				When("publisher throws error while publishing messages to temporary queue", func() {
					BeforeEach(func() {
						selectorMock.On(util.NameOf(selectorMock.IsSelected), mock.Anything).Unset() // because of error while publishing, we don't want to call isSelected N times
						selectorMock.On(util.NameOf(selectorMock.IsSelected), mock.Anything).Return(true, nil).Once()

						handler.On(util.NameOf(handler.Handle), mock.Anything).Unset() // because of error while publishing, we don't want to call handle N times
						handler.On(util.NameOf(handler.Handle), mock.Anything).Return(true, nil).Once()

						pubMock.On(util.NameOf(pubMock.Publish), "tempQueue", mock.Anything).Return(errors.New("")).Once()
					})

					When("acknowledger fails to reject delivery", func() {
						BeforeEach(func() {
							ackMock.On(util.NameOf(ackMock.Reject), srcMessages[0].DeliveryTag, true).Return(errors.New(""))
						})

						It("returns error", func() {
							err := manager.Manage(context.Background(), "srcQueue")
							Expect(err).To(HaveOccurred())
							for _, msg := range srcMessages {
								Expect(ackMock.AckedTags[msg.DeliveryTag]).ToNot(BeTrue())
							}
						})
					})

					When("acknowledger succeeds to reject delivery", func() {
						BeforeEach(func() {
							ackMock.On(util.NameOf(ackMock.Reject), srcMessages[0].DeliveryTag, true).Return(nil)
						})

						It("returns error", func() {
							err := manager.Manage(context.Background(), "srcQueue")
							Expect(err).To(HaveOccurred())
							for _, msg := range srcMessages {
								Expect(ackMock.AckedTags[msg.DeliveryTag]).ToNot(BeTrue())
							}
						})
					})
				})

				When("publisher succeeds to publish all messages to temporary queue", func() {

					var tempMessages []amqp091.Delivery

					BeforeEach(func() {
						tempMessages = []amqp091.Delivery{
							{DeliveryTag: sequenceNumber.Add(1), Acknowledger: ackMock},
							{DeliveryTag: sequenceNumber.Add(1), Acknowledger: ackMock},
							{DeliveryTag: sequenceNumber.Add(1), Acknowledger: ackMock},
						}
						pubMock.On(util.NameOf(pubMock.Publish), "tempQueue", mock.Anything).Return(nil).Times(len(srcMessages))
					})

					When("acknowledger fails to ack delivery", func() {
						BeforeEach(func() {
							ackMock.On(util.NameOf(ackMock.Ack), mock.Anything, false).Return(errors.New(""))

							selectorMock.On(util.NameOf(selectorMock.IsSelected), mock.Anything).Unset() // because of error while publishing, we don't want to call isSelected N times
							selectorMock.On(util.NameOf(selectorMock.IsSelected), mock.Anything).Return(true, nil).Once()

							handler.On(util.NameOf(handler.Handle), mock.Anything).Unset() // because of error while acking, we don't want to call handle N times
							handler.On(util.NameOf(handler.Handle), mock.Anything).Return(true, nil).Once()

							pubMock.On(util.NameOf(pubMock.Publish), "tempQueue", mock.Anything).Unset() // because of error while acking, we don't want to call publish N times
							pubMock.On(util.NameOf(pubMock.Publish), "tempQueue", mock.Anything).Return(nil).Once()
						})

						It("returns error", func() {
							err := manager.Manage(context.Background(), "srcQueue")
							Expect(err).To(HaveOccurred())
						})
					})

					When("acknowledger succeeds to ack deliveries", func() {
						BeforeEach(func() {
							for _, msg := range srcMessages {
								ackMock.On(util.NameOf(ackMock.Ack), msg.DeliveryTag, false).Return(nil)
							}
						})

						It("throws error if context is eventually cancelled", func() {
							ctx, cancel := context.WithCancel(context.Background())
							go func() {
								time.Sleep(time.Millisecond * 500)
								cancel()
							}()
							err := manager.Manage(ctx, "srcQueue")
							Expect(err).To(HaveOccurred())
						})

						When("moving messages back to the source queue", func() {
							BeforeEach(func() {
								tempQueue := initReadChannel(tempMessages)
								conMock.On(util.NameOf(conMock.Consume), "tempQueue").Return(tempQueue, nil).Once()
							})

							When("publisher throws error while publishing messages back to the source queue", func() {
								BeforeEach(func() {
									pubMock.On(util.NameOf(pubMock.Publish), "srcQueue", mock.Anything).Return(errors.New("")).Once()
								})

								When("acknowledger fails to reject delivery", func() {
									BeforeEach(func() {
										ackMock.On(util.NameOf(ackMock.Reject), tempMessages[0].DeliveryTag, false).Return(errors.New(""))
									})

									It("throws error", func() {
										err := manager.Manage(context.Background(), "srcQueue")
										Expect(err).To(HaveOccurred())

										for i := 0; i < len(srcMessages); i++ {
											Expect(ackMock.AckedTags[srcMessages[i].DeliveryTag]).To(BeTrue())
											Expect(ackMock.AckedTags[tempMessages[i].DeliveryTag]).ToNot(BeTrue())
										}
									})
								})

								When("acknowledger succeeds to reject delivery", func() {
									BeforeEach(func() {
										ackMock.On(util.NameOf(ackMock.Reject), tempMessages[0].DeliveryTag, false).Return(nil)
									})

									It("throws error", func() {
										err := manager.Manage(context.Background(), "srcQueue")
										Expect(err).To(HaveOccurred())

										for i := 0; i < len(srcMessages); i++ {
											Expect(ackMock.AckedTags[srcMessages[i].DeliveryTag]).To(BeTrue())
											Expect(ackMock.AckedTags[tempMessages[i].DeliveryTag]).ToNot(BeTrue())
										}
									})
								})
							})

							When("publisher succeeds to publish all messages back to the source queue", func() {
								BeforeEach(func() {
									pubMock.On(util.NameOf(pubMock.Publish), "srcQueue", mock.Anything).Return(nil).Times(len(tempMessages))
								})

								When("publisher fails to ack delivery", func() {
									BeforeEach(func() {
										pubMock.On(util.NameOf(pubMock.Publish), "srcQueue", mock.Anything).Unset() // because of error while acking, we don't want to call Publish N times
										pubMock.On(util.NameOf(pubMock.Publish), "srcQueue", mock.Anything).Return(nil).Once()
										ackMock.On(util.NameOf(ackMock.Ack), tempMessages[0].DeliveryTag, false).Return(errors.New(""))
									})

									It("throws error", func() {
										err := manager.Manage(context.Background(), "srcQueue")
										Expect(err).To(HaveOccurred())
									})
								})

								When("publisher succeeds to ack deliveries", func() {
									BeforeEach(func() {
										for _, msg := range tempMessages {
											ackMock.On(util.NameOf(ackMock.Ack), msg.DeliveryTag, false).Return(nil)
										}
									})

									It("acks all messages in the source queue and temporary queue and succeeds", func() {
										err := manager.Manage(context.Background(), "srcQueue")
										Expect(err).ToNot(HaveOccurred())

										for i := 0; i < len(srcMessages); i++ {
											Expect(ackMock.AckedTags[srcMessages[i].DeliveryTag]).To(BeTrue())
											Expect(ackMock.AckedTags[tempMessages[i].DeliveryTag]).To(BeTrue())
										}
									})
								})
							})
						})
					})
				})
			})
		})
	})

	When("no messages", func() {
		BeforeEach(func() {
			srcQueue := make(<-chan amqp091.Delivery)
			tempQueue := make(<-chan amqp091.Delivery)
			conMock.On(util.NameOf(conMock.Consume), "srcQueue").Return(srcQueue, nil).Once()
			conMock.On(util.NameOf(conMock.Consume), "tempQueue").Return(tempQueue, nil).Once()
		})

		It("doesn't throw error", func() {
			err := manager.Manage(context.Background(), "srcQueue")
			Expect(err).ToNot(HaveOccurred())
		})

		When("context is cancelled", func() {
			BeforeEach(func() {
				conMock.On(util.NameOf(conMock.Consume), "tempQueue").Unset()
			})

			It("throws error", func() {
				ctx, cancel := context.WithCancel(context.Background())
				go func() {
					time.Sleep(time.Millisecond * 500)
					cancel()
				}()
				err := manager.Manage(ctx, "srcQueue")
				Expect(err).To(HaveOccurred())
			})
		})
	})
})
